package logging

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/utils/wordwrap"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"time"
)

const logFile = "thelma.log"
const configPrefix = "logging"

// globalWriter is the writer that zerolog's global log.Logger is configured to write to.
// We track it in a package-level variable so that WithMask can wrap it with a masking writer.
// Initialized with a basic writer here that is overwritten with a more complex/configurable writer during InitializeLogging
var globalWriter = NewMaskingWriter(zerolog.ConsoleWriter{Out: os.Stderr})

// logConfig is a configuration struct for the logging package
type logConfig struct {
	Console struct {
		// Log level for console messages
		Level string `default:"info" validate:"oneof=trace debug info warn error"`
		// WordWrap if true, wrap long log lines at word boundary to max terminal width
		WordWrap bool `default:"true"`
	}
	File struct {
		// Log
		Enabled   bool   `default:"true"`
		Dir       string // Default is $THELMA_ROOT/log
		Level     string `default:"debug" validate:"oneof=trace debug info warn error"`
		KeepFiles int    `default:"5" validate:"gte=0"`
		MaxSizeMb int    `default:"8" validate:"gte=0"`
	}
	Caller struct {
		// Set to true to include caller information (source file and line number) in log messages.
		Enabled bool `default:"false"`
	}
}

func init() {
	Bootstrap()
}

// Bootstrap configure global zerolog logger with a basic console logger
// to catch any messages that are logged before full Thelma initialization
func Bootstrap() {
	log.Logger = log.Output(globalWriter).Level(zerolog.DebugLevel)
}

// MaskSecret configures the global logger to mask the given secret(s)
func MaskSecret(secret ...string) {
	globalWriter.MaskSecrets(secret...)
}

// Initialize updates the global Zerolog logger to match Thelma's configuration.
// It should be called once during Thelma initialization.
func Initialize(thelmaConfig config.Config, thelmaRoot root.Root) error {
	cfg, err := loadConfig(thelmaConfig, thelmaRoot)
	if err != nil {
		return fmt.Errorf("logging initialization failed: %v", err)
	}

	// init new logger based on configuration
	writer, err := newCompositeWriter(cfg, os.Stderr)
	if err != nil {
		return fmt.Errorf("logging initialization failed: %v", err)
	}

	logger, err := newLogger(cfg, writer)
	if err != nil {
		return fmt.Errorf("logging initialization failed: %v", err)
	}

	// replace default Zerolog logger with our custom logger
	globalWriter = writer
	log.Logger = *logger

	return nil
}

// Initialize a logConfig based on given thelmaConfig
func loadConfig(thelmaConfig config.Config, thelmaRoot root.Root) (*logConfig, error) {
	cfg := &logConfig{}

	if err := thelmaConfig.Unmarshal(configPrefix, cfg); err != nil {
		return nil, err
	}

	// Set dynamic defaults
	if cfg.File.Dir == "" { // Default log dir to ~/.thelma/logs
		cfg.File.Dir = thelmaRoot.LogDir()
	}

	return cfg, nil
}

// Construct a new composite console + file writer based on supplied configuration, wrapped in a MaskingWriter for redacting secrets
func newCompositeWriter(cfg *logConfig, consoleStream io.Writer) (*MaskingWriter, error) {
	var writers []io.Writer

	// Create console writer
	writers = append(writers, newConsoleWriter(cfg, consoleStream))

	// Create file writer
	if cfg.File.Enabled {
		fw, err := newFileWriter(cfg)
		if err != nil {
			return nil, err
		}
		writers = append(writers, fw)
	}

	// Combine writers into a multi writer
	combined := zerolog.MultiLevelWriter(writers...)

	return NewMaskingWriter(combined), nil
}

// Construct a new logger based on supplied configuration
func newLogger(cfg *logConfig, compositeWriter zerolog.LevelWriter) (*zerolog.Logger, error) {
	ctx := zerolog.New(compositeWriter).With()

	// If enabled, include source file / line number in log messages
	if cfg.Caller.Enabled {
		ctx = ctx.Caller()
	}

	// Add timestamps to logs
	ctx = ctx.Timestamp()

	logger := ctx.Logger()
	return &logger, nil
}

func newConsoleWriter(cfg *logConfig, consoleStream io.Writer) zerolog.LevelWriter {
	outputStream := consoleStream

	if cfg.Console.WordWrap {
		outputStream = NewWrappingWriter(consoleStream, func(options *wordwrap.Options) {
			options.DynamicMaxWidth = true
			options.EscapeNewlineStringLiteral = true

			// try to match the width of zero-log's preconfigured date formatter
			options.Padding = "           "
			if time.Now().Hour()%12 >= 10 {
				options.Padding += " "
			}
		})
	}

	consoleWriter := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = outputStream
	})

	return NewFilteredWriter(zerolog.MultiLevelWriter(consoleWriter), parseLogLevel(cfg.Console.Level))
}

func newFileWriter(cfg *logConfig) (zerolog.LevelWriter, error) {
	if err := os.MkdirAll(cfg.File.Dir, 0700); err != nil {
		return nil, fmt.Errorf("error creating log directory %s: %v", cfg.File.Dir, err)
	}
	rollingWriter := &lumberjack.Logger{
		Filename:   path.Join(cfg.File.Dir, logFile),
		MaxSize:    cfg.File.MaxSizeMb,
		MaxBackups: cfg.File.KeepFiles,
	}
	return NewFilteredWriter(zerolog.MultiLevelWriter(rollingWriter), parseLogLevel(cfg.File.Level)), nil
}

// parseLogLevel parse log level string to zerolog.Level
func parseLogLevel(levelStr string) zerolog.Level {
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to parse log level %q: %v", levelStr, err)
		return zerolog.InfoLevel
	}
	return level
}
