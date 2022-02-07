package testutils

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

var alphaNumeric = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Args convenience function to generate tokenized argument list from format string w/ args
//
// Eg. args("-e   %s", "dev") -> []string{"-e", "dev"}
func Args(format string, a ...interface{}) []string {
	formatted := fmt.Sprintf(format, a...)
	return strings.Fields(formatted)
}

// Cwd convenience function to return current working directory
func Cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// RandString generates a random alphanumeric string (a-z0-9) of length n
func RandString(n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = alphaNumeric[rand.Intn(len(alphaNumeric))]
	}
	return string(result)
}
