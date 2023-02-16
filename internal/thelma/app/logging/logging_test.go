package logging

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

const kbBytes = 1024
const mbBytes = 1024 * kbBytes

type testMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Caller  string `json:"caller"`
}

type testResult struct {
	fileMessages    []testMessage
	consoleMessages []string
	logDir          string
	logFile         string
}

func Test_newLogger(t *testing.T) {
	testCases := []struct {
		name           string
		thelmaSettings map[string]interface{}
		maskSecrets    []string
		setupFn        func(*zerolog.Logger)
		verifyFn       func(t *testing.T, r testResult)
	}{
		{
			name: "with default settings, info messages should be logged to both places",
			setupFn: func(logger *zerolog.Logger) {
				logger.Info().Msg("hello world")
			},
			verifyFn: func(t *testing.T, r testResult) {
				assert.Equal(t, 1, len(r.consoleMessages))
				assert.Regexp(t, "INF.*hello world", r.consoleMessages[0])

				assert.Equal(t, 1, len(r.fileMessages))
				assert.Equal(t, "hello world", r.fileMessages[0].Message)
				assert.Equal(t, "info", r.fileMessages[0].Level)
				assert.Equal(t, "", r.fileMessages[0].Caller)
			},
		},
		{
			name: "with default settings, debug messages should go to file but not console",
			setupFn: func(logger *zerolog.Logger) {
				logger.Debug().Msg("hello world")
			},
			verifyFn: func(t *testing.T, r testResult) {
				assert.Equal(t, 0, len(r.consoleMessages))

				assert.Equal(t, 1, len(r.fileMessages))
				assert.Equal(t, "hello world", r.fileMessages[0].Message)
				assert.Equal(t, "debug", r.fileMessages[0].Level)
				assert.Equal(t, "", r.fileMessages[0].Caller)
			},
		},
		{
			name: "when file logging is disabled, no log messages should be written",
			thelmaSettings: map[string]interface{}{
				"logging.file.enabled": "false",
			},
			setupFn: func(logger *zerolog.Logger) {
				logger.Info().Msg("hello world")
			},
			verifyFn: func(t *testing.T, r testResult) {
				assert.Equal(t, 1, len(r.consoleMessages))

				assert.NoFileExists(t, r.logFile)
				assert.Equal(t, 0, len(r.fileMessages))
			},
		},
		{
			name: "caller=true should add caller info to log",
			thelmaSettings: map[string]interface{}{
				"logging.caller.enabled": "true",
			},
			setupFn: func(logger *zerolog.Logger) {
				logger.Info().Msg("hello world")
			},
			verifyFn: func(t *testing.T, r testResult) {
				assert.Equal(t, 1, len(r.consoleMessages))
				assert.Regexp(t, "logging_test.go", r.consoleMessages[0])

				assert.Equal(t, 1, len(r.fileMessages))
				assert.Regexp(t, "logging_test.go", r.fileMessages[0].Caller)
			},
		},
		{
			name: "log files should rotate automatically",
			thelmaSettings: map[string]interface{}{
				"logging.file.keepfiles": 2,
				"logging.file.maxsizemb": 1,
			},
			setupFn: func(logger *zerolog.Logger) {
				numMessages := 3 * mbBytes / 10 // Our message is way longer than 10 bytes, so this will rotate more than 3 times
				log.Info().Msgf("Writing %d log messages", numMessages)
				start := time.Now()
				for i := 0; i < numMessages; i++ {
					logger.Debug().Msgf("this is a test message: %d", i)
				}
				log.Info().Msgf("Finished writing %d messages in %s", numMessages, time.Since(start))
			},
			verifyFn: func(t *testing.T, r testResult) {
				assert.Equal(t, 0, len(r.consoleMessages))

				entries, err := os.ReadDir(r.logDir)
				if !assert.NoError(t, err) {
					return
				}

				var files []os.FileInfo
				for _, entry := range entries {
					fileInfo, err := entry.Info()
					assert.NoError(t, err)
					log.Debug().Msgf("Generated log file: %v %v", fileInfo.Name(), fileInfo.Size())
					files = append(files, fileInfo)
				}

				assert.Equal(t, 3, len(entries), "Expected 3 log entries to be generated")
				assert.FileExists(t, r.logFile)
				for _, f := range files {
					if path.Base(f.Name()) != logFile {
						assert.InDelta(t, mbBytes, f.Size(), kbBytes, "Expected rotated log file to be within 1kb delta of 1mb, but log file %s has size %d", f.Name(), f.Size())
					}
				}
			},
		},
		{
			name:        "with mask should mask multiple secrets in messages",
			maskSecrets: []string{"foo", "bar", "baz"},
			setupFn: func(logger *zerolog.Logger) {
				logger.Info().Msg("Here is a long message with secrets: foo, also bar, and baz")
			},
			verifyFn: func(t *testing.T, r testResult) {
				assert.Equal(t, 1, len(r.consoleMessages))
				assert.Regexp(t, `INF.*Here is a long message with secrets: \*\*\*\*\*\*, also \*\*\*\*\*\*, and \*\*\*\*\*\*`, r.consoleMessages[0])

				assert.Equal(t, 1, len(r.fileMessages))
				assert.Equal(t, "Here is a long message with secrets: ******, also ******, and ******", r.fileMessages[0].Message)
			},
		},
		{
			name:        "with mask should mask secrets in contextual fields",
			maskSecrets: []string{"foo"},
			setupFn: func(logger *zerolog.Logger) {
				_logger := logger.With().
					Str("key1", "foo").
					Str("key2", "extra foo stuff").
					Logger()
				_logger.Info().Msg("Hello foo")
			},
			verifyFn: func(t *testing.T, r testResult) {
				assert.Equal(t, 1, len(r.consoleMessages))
				assert.Regexp(t, `INF.*Hello \*\*\*\*\*\*.*key1.*=.*\*\*\*\*\*\*.*key2.*=.*"extra \*\*\*\*\*\* stuff"`, r.consoleMessages[0])

				jsonLog, err := os.ReadFile(r.logFile)
				require.NoError(t, err)
				assert.Regexp(t, `{"level":"info","key1":"\*\*\*\*\*\*","key2":"extra \*\*\*\*\*\* stuff","time":".*","message":"Hello \*\*\*\*\*\*"}`, string(jsonLog))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logDir := t.TempDir()

			settings := map[string]interface{}{
				"logging.file.dir": logDir,
			}
			if tc.thelmaSettings != nil {
				for k, v := range tc.thelmaSettings {
					settings[k] = v
				}
			}

			thelmaConfig, err := config.NewTestConfig(t, settings)
			require.NoError(t, err)

			cfg, err := loadConfig(thelmaConfig, root.NewAt(t.TempDir()))
			require.NoError(t, err)

			fakeConsoleWriter := &bytes.Buffer{}
			maskingWriter, err := newCompositeWriter(cfg, fakeConsoleWriter)
			require.NoError(t, err)

			if len(tc.maskSecrets) > 0 {
				maskingWriter.MaskSecrets(tc.maskSecrets...)
			}

			logger, err := newLogger(cfg, maskingWriter)
			require.NoError(t, err)

			if tc.setupFn != nil {
				tc.setupFn(logger)
			}

			if tc.verifyFn != nil {
				consoleMessages := parseConsoleLog(fakeConsoleWriter)

				file := path.Join(logDir, logFile)
				fileMessages, err := parseLogFile(file)
				if !assert.NoError(t, err) {
					return
				}

				tc.verifyFn(t, testResult{fileMessages, consoleMessages, logDir, file})
			}
		})
	}
}

func parseConsoleLog(buf *bytes.Buffer) []string {
	var lines []string
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	return lines
}

// parse newline-separated JSON log messages into a slice of structs
func parseLogFile(logFile string) ([]testMessage, error) {
	var messages []testMessage

	if _, err := os.Stat(logFile); err != nil {
		if os.IsNotExist(err) {
			log.Warn().Msgf("log file %s does not exist, returning empty message list", logFile)
			return messages, nil
		} else {
			return nil, fmt.Errorf("error reading log file %s: %v", logFile, err)
		}
	}

	f, err := os.Open(logFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Msgf("error closing test log file %s: %v", logFile, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	lineNumber := 1
	for scanner.Scan() {
		var m testMessage
		content := scanner.Bytes()
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			return nil, fmt.Errorf("error parsing message on line %d of %s as JSON: %q", lineNumber, logFile, string(content))
		}
		messages = append(messages, m)
		lineNumber++
	}

	return messages, nil
}
