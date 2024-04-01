package environmentflags

import (
	. "github.com/broadinstitute/thelma/internal/thelma/utils/testutils"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var environmentVariablesToUnset []string
var environmentVariablesToRestore [][]string

func temporarilySetEnvVar(key, value string) error {
	if existingValue, present := os.LookupEnv(key); present {
		environmentVariablesToRestore = append(environmentVariablesToRestore, []string{key, existingValue})
	} else {
		environmentVariablesToUnset = append(environmentVariablesToUnset, key)
	}
	return os.Setenv(key, value)
}

func restoreEnvVars() error {
	for _, pair := range environmentVariablesToRestore {
		if err := os.Setenv(pair[0], pair[1]); err != nil {
			return err
		}
	}
	environmentVariablesToRestore = nil
	for _, key := range environmentVariablesToUnset {
		if err := os.Unsetenv(key); err != nil {
			return err
		}
	}
	environmentVariablesToUnset = nil
	return nil
}

func TestSetFlagsFromEnvironment(t *testing.T) {

	testCases := []struct {
		name         string
		args         []string
		envVars      map[string]string
		expectErrors []string
		expectString string
		expectBool   bool
		expectInt    int
		expectSlice  []string
	}{
		{
			name:         "base case",
			args:         []string{},
			envVars:      map[string]string{},
			expectErrors: nil,
			expectString: "default",
			expectBool:   true,
			expectInt:    1,
			expectSlice:  []string{"list"},
		},
		{
			name:         "set all on command line",
			args:         Args("--test-string=blah --test-bool=false --test-int=2 --test-string-slice=foo,bar"),
			envVars:      map[string]string{},
			expectErrors: nil,
			expectString: "blah",
			expectBool:   false,
			expectInt:    2,
			expectSlice:  []string{"foo", "bar"},
		},
		{
			name: "set all in environment",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_STRING":       "blah",
				"THELMA_TEST_PARAM_TEST_BOOL":         "false",
				"THELMA_TEST_PARAM_TEST_INT":          "2",
				"THELMA_TEST_PARAM_TEST_STRING_SLICE": "foo,bar",
			},
			expectErrors: nil,
			expectString: "blah",
			expectBool:   false,
			expectInt:    2,
			expectSlice:  []string{"foo", "bar"},
		},
		{
			name: "cli takes precedence over environment",
			args: Args(" --test-string=blah --test-bool=false --test-int=2 --test-string-slice=foo,bar --%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_STRING":       "notblah",
				"THELMA_TEST_PARAM_TEST_BOOL":         "true",
				"THELMA_TEST_PARAM_TEST_INT":          "3",
				"THELMA_TEST_PARAM_TEST_STRING_SLICE": "baz",
			},
			expectErrors: nil,
			expectString: "blah",
			expectBool:   false,
			expectInt:    2,
			expectSlice:  []string{"foo", "bar"},
		},
		{
			name: "mixed",
			args: Args("--test-string=blah --test-bool=false --%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_BOOL": "true",
				"THELMA_TEST_PARAM_TEST_INT":  "2",
			},
			expectErrors: nil,
			expectString: "blah",           // CLI, uncontested
			expectBool:   false,            // CLI, precedence over env
			expectInt:    2,                // env, uncontested
			expectSlice:  []string{"list"}, // default
		},
		{
			name: "ignores spurious env vars",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_SOME_OTHER_PREFIX_STRING": "blah", // wrong prefix
				"THELMA_TEST_PARAM_SOME_FLAG":          "true", // flag doesn't exist
				"THELMA_TEST_PARAM_test-string":        "blah", // wrong name transform
				"THELMA_TEST_PARAM_test_string":        "blah", // wrong name transform
				"THELMA_TEST_PARAM_TEST-STRING":        "blah", // wrong name transform
			},
			expectErrors: nil,
			expectString: "default",
			expectBool:   true,
			expectInt:    1,
			expectSlice:  []string{"list"},
		},
		{
			name: "errors upon bool type coercion",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_BOOL": "not a bool",
			},
			expectErrors: []string{
				"failed to set --test-bool from environment variable THELMA_TEST_PARAM_TEST_BOOL with value `not a bool`: strconv.ParseBool: parsing \"not a bool\": invalid syntax",
			},
		},
		{
			name: "errors upon bool type coercion (empty)",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_BOOL": "",
			},
			expectErrors: []string{
				"failed to set --test-bool from environment variable THELMA_TEST_PARAM_TEST_BOOL with value ``: strconv.ParseBool: parsing \"\": invalid syntax",
			},
		},
		{
			name: "errors upon int type coercion",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_INT": "not an int",
			},
			expectErrors: []string{
				"failed to set --test-int from environment variable THELMA_TEST_PARAM_TEST_INT with value `not an int`: strconv.ParseInt: parsing \"not an int\": invalid syntax",
			},
		},
		{
			name: "errors upon int type coercion (empty)",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_INT": "",
			},
			expectErrors: []string{
				"failed to set --test-int from environment variable THELMA_TEST_PARAM_TEST_INT with value ``: strconv.ParseInt: parsing \"\": invalid syntax",
			},
		},
		{
			name: "string accepts empty",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_STRING": "",
			},
			expectErrors: nil,
			expectString: "",
			expectBool:   true,
			expectInt:    1,
			expectSlice:  []string{"list"},
		},
		{
			name: "string slice accepts empty",
			args: Args("--%s=THELMA_TEST_PARAM_", flagsFromEnvironmentPrefixFlag),
			envVars: map[string]string{
				"THELMA_TEST_PARAM_TEST_STRING_SLICE": ",,,",
			},
			expectErrors: nil,
			expectString: "default",
			expectBool:   true,
			expectInt:    1,
			expectSlice:  []string{"", "", "", ""},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.PanicOnError)

			AddFlag(flags)
			flags.String("test-string", "default", "test string flag")
			flags.Bool("test-bool", true, "test bool flag")
			flags.Int("test-int", 1, "test int flag")
			flags.StringSlice("test-string-slice", []string{"list"}, "test string slice flag")

			assert.NoError(t, flags.Parse(testCase.args))

			for key, value := range testCase.envVars {
				assert.NoError(t, temporarilySetEnvVar(key, value))
			}

			gotErrors := SetFlagsFromEnvironment(flags)

			assert.NoError(t, restoreEnvVars())

			var gotErrorStrings []string
			for _, err := range gotErrors {
				gotErrorStrings = append(gotErrorStrings, err.Error())
			}
			assert.Equal(t, testCase.expectErrors, gotErrorStrings)

			if len(gotErrors) > 0 {
				return
			}

			gotString, err := flags.GetString("test-string")
			assert.Equal(t, testCase.expectString, gotString)
			assert.NoError(t, err)
			gotBool, err := flags.GetBool("test-bool")
			assert.Equal(t, testCase.expectBool, gotBool)
			assert.NoError(t, err)
			gotInt, err := flags.GetInt("test-int")
			assert.Equal(t, testCase.expectInt, gotInt)
			assert.NoError(t, err)
			gotSlice, err := flags.GetStringSlice("test-string-slice")
			assert.Equal(t, testCase.expectSlice, gotSlice)
			assert.NoError(t, err)
		})
	}
}
