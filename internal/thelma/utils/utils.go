// Package utils contains miscellaneous utility code
package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mcuadros/go-defaults"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"

	"github.com/mattn/go-isatty"
)

// Interactive returns true if Thelma is running in an interactive shell, false otherwise. Useful for detecting
// if Thelma is running in CI pipelines or on a dev laptop
func Interactive() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

// TerminalWidth returns width of the current interactive terminal. Returns 0 if shell is not interactive
// or if terminal width could not be detected
func TerminalWidth() int {
	if !Interactive() {
		return 0
	}
	width, _, err := term.GetSize(0)
	if err != nil {
		return 0
	}
	return width
}

// ExpandAndVerifyExists Expand relative path to absolute, and make sure it exists.
// This is necessary for many arguments because Helmfile assumes paths
// are relative to helmfile.yaml and we want them to be relative to CWD.
func ExpandAndVerifyExists(filePath string, description string) (string, error) {
	expanded, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(expanded); os.IsNotExist(err) {
		return "", errors.Errorf("%s does not exist: %s", description, expanded)
	} else if err != nil {
		return "", errors.Errorf("error reading %s %s: %v", description, expanded, err)
	}

	return expanded, nil
}

// IsIPV4Address returns true if addr is a valid ipv4 address
func IsIPV4Address(addr string) bool {
	ip := net.ParseIP(addr)

	return ip != nil && ip.To4() != nil
}

// QuoteJoin quotes all strings in a slice and joins them with `, `
// eg.
// QuoteJoin([]string{`a`, `b`, `c`}, `, `)
// ->
// `"a", "b", "c"`
func QuoteJoin(strs []string) string {
	var quoted []string
	for _, s := range strs {
		quoted = append(quoted, fmt.Sprintf("%q", s))
	}
	return strings.Join(quoted, ", ")
}

// FileExists returns true if the file exists, false otherwise, and an error if an error occurs
func FileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Nullable is a utility to turn a value into a pointer to that value
func Nullable[T any](val T) *T {
	return &val
}

// CloseWarn gracefully handle Close() error when a prior error is more salient
func CloseWarn(closer io.Closer, err error) error {
	closeErr := closer.Close()
	if err == nil {
		return closeErr
	}
	if closeErr == nil {
		return err
	}
	log.Error().Err(closeErr).Msg("close error")
	return err
}

// PathToRunningThelmaExecutable returns the path to the currently-running
// Thelma binary executable.
// Note that this could be _outside_ Thelma's configured root directory
// (i.e., not ~/.thelma/releases/current/bin).
// For example:
//   - During initial installation, Thelma is run out of Thelma release archive
//     that is unpacked into a temp directory.
//   - In CI pipelines, Thelma is run out of a well-known path on its Docker image
//     /thelma/bin/thelma
//   - When Thelma is built locally during development, it is run out of the build
//     output directory, ./output/bin/thelma
//
// Also note that this might not be a running Thelma binary at all, for example
// if this code is executed from a unit test (command will be the `go` executable)
func PathToRunningThelmaExecutable() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", errors.Errorf("error finding path to currently running executable: %v", err)
	}

	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return "", errors.Errorf("error finding path to currently running executable: %v", err)
	}

	return executable, nil
}

// CollateOptionsWithDefaults given a struct annotated with `default` annotations from
// https://github.com/mcuadros/go-defaults, set defaults and call CollateOptions
// set utility for collating option functions into a struct
func CollateOptionsWithDefaults[T any](optFns ...func(*T)) T {
	var options T
	defaults.SetDefaults(&options)
	return CollateOptions(options, optFns...)
}

// CollateOptions utility for collating option functions into a struct
func CollateOptions[T any](defaults T, optFns ...func(*T)) T {
	for _, optFn := range optFns {
		if optFn == nil {
			continue
		}
		optFn(&defaults)
	}
	return defaults
}

// Not negates a single-argument predicate
func Not[T any](fn func(T) bool) func(T) bool {
	return func(t T) bool {
		return !fn(t)
	}
}

func UnsetOrEmpty[T comparable](val *T) bool {
	var t T
	return val == nil || *val == t
}

// JoinSelector join map of label key-value pairs {"a":"b", "c":"d"} into selector string "a=b,c=d"
func JoinSelector(labels map[string]string) string {
	var list []string
	for name, value := range labels {
		list = append(list, fmt.Sprintf("%s=%s", name, value))
	}
	sort.Strings(list)
	return strings.Join(list, ",")
}

func OverrideEnvVarForTest(t *testing.T, variable string, value string, test func()) {
	original, originalPresent := os.LookupEnv(variable)
	err := os.Setenv(variable, value)
	if err != nil {
		t.Fatalf("error setting %s: %v", variable, err)
	}
	test()
	if originalPresent {
		err = os.Setenv(variable, original)
	} else {
		err = os.Unsetenv(variable)
	}
	if err != nil {
		t.Fatalf("error resetting %s: %v", variable, err)
	}
}
