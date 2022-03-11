package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

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
	return strings.Join(strs, ", ")
}
