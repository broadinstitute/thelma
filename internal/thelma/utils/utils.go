package utils

import (
	"fmt"
	"os"
	"path/filepath"
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

// AsInterfaces is for converting slices of a given type to a slice of interfaces.
func AsInterfaces(a ...interface{}) []interface{} {
	res := make([]interface{}, len(a))
	for i := range a {
		res[i] = a[i]
	}
	return res
}
