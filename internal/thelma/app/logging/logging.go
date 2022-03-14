package logging

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
)

const defaultDir = "logs"
const logFile = "thelma.log"
const configPrefix = "logging"

// logConfig is a configuration struct for the logging package
type logConfig struct {
	Console struct {
		// Log level for console messages
		Level string `default:"info" validate:"oneof=trace debug info warn error"`
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

func (cfg *logConfig) logDir() string {
	dir := cfg.File.Dir
	if dir == "" {
		// default log dir is ~/.thelma/logs
		dir = path.Join(root.Dir(), defaultDir)
	}
	return dir
}

// Bootstrap configure global zerolog logger with a basic console logger
// to catch any messages that are logged before full Thelma initialization
func Bootstrap() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// InitializeLogging updates the global Zerolog logger to match Thelma's configuration.
// It should be called once during Thelma initialization.
func InitializeLogging(thelmaConfig config.Config) error {
	cfg, err := loadConfig(thelmaConfig)
	if err != nil {
		return fmt.Errorf("logging initialization failed: %v", err)
	}

	// init new logger based on configuration
	logger, err := newLogger(cfg, os.Stderr)
	if err != nil {
		return fmt.Errorf("logging initialization failed: %v", err)
	}

	// replace default Zerolog logger with our custom logger
	log.Logger = *logger
	return nil
}

// Initialize a logConfig based on given thelmaConfig
func loadConfig(thelmaConfig config.Config) (*logConfig, error) {
	cfg := &logConfig{}

	if err := thelmaConfig.Unmarshal(configPrefix, cfg); err != nil {
		return nil, err
	}

	// Set dynamic defaults
	if cfg.File.Dir == "" { // Default log dir to ~/.thelma/logs
		cfg.File.Dir = path.Join(root.Dir(), defaultDir)
	}

	return cfg, nil
}

// Construct a new logger based on supplied configuration
func newLogger(cfg *logConfig, consoleStream io.Writer) (*zerolog.Logger, error) {
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
	multi := zerolog.MultiLevelWriter(writers...)

	ctx := zerolog.New(multi).With()

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
	writer := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = consoleStream
	})
	return NewFilteredWriter(zerolog.MultiLevelWriter(writer), parseLogLevel(cfg.Console.Level))
}

func newFileWriter(cfg *logConfig) (zerolog.LevelWriter, error) {
	if err := os.MkdirAll(cfg.logDir(), 0700); err != nil {
		return nil, fmt.Errorf("error creating log directory %s: %v", cfg.logDir(), err)
	}
	rollingWriter := &lumberjack.Logger{
		Filename:   path.Join(cfg.logDir(), logFile),
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
