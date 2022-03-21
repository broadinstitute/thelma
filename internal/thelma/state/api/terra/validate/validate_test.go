package validate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnvironmentName(t *testing.T) {
	cases := map[string]bool{
		"":                                  false,
		"a":                                 true,
		"1":                                 false,
		"-":                                 false,
		"aa":                                true,
		"a-":                                false,
		"-a":                                false,
		"a-a":                               true,
		"a-a-a-a-a":                         true,
		"a1":                                true,
		"a-100":                             true,
		"a9-23-23fa-fjdf-2":                 true,
		"1a":                                false,
		"a-1-2-3-4":                         true,
		"con--secutive-dashes":              false,
		"name-with-thirty-one-characters":   true,
		"name-with-thirty-two2-characters":  true,
		"name-with-thirty-three-characters": false,
	}

	for name, ok := range cases {
		err := EnvironmentName(name)
		if ok {
			assert.NoError(t, err, "%q should be permitted", name)
		} else {
			assert.Error(t, err, "%q should NOT permitted", name)
		}
	}
}
