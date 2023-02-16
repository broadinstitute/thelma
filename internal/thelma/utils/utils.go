// Package utils contains miscellaneous utility code
package utils

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
)

// Interactive returns true if Thelma is running in an interactive shell, false otherwise. Useful for detecting
// if Thelma is running in CI pipelines or on a dev laptop
func Interactive() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
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
		return "", fmt.Errorf("%s does not exist: %s", description, expanded)
	} else if err != nil {
		return "", fmt.Errorf("error reading %s %s: %v", description, expanded, err)
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
