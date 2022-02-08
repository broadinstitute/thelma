package config

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/testutils"
	"github.com/mcuadros/go-defaults"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fake config struct for testing unmarshalling.
type testConfig struct {
	Logging struct {
		Console struct {
			Level string `default:"info" validate:"oneof=trace debug info warn error"`
		}
		File struct {
			Enabled   bool   `default:"true"`
			Level     string `default:"debug" validate:"oneof=trace debug info warn error"`
			KeepFiles int    `default:"5" validate:"gte=0"`
			MaxSizeMb int    `default:"8" validate:"gte=0"`
		}
	}
	Docker struct {
		HostAliases []HostAlias
	}
	Airports struct {
		IataCodes map[string]Airport
	}
}

type HostAlias struct {
	Name string `default:"localhost" validate:"hostname"`
	Addr string `default:"127.0.0.1" validate:"ip"`
}

type Airport struct {
	Name     string
	Location Location
}

type Location struct {
	City    string
	Country string `validate:"iso3166_1_alpha2"`
}

func TestConfig_Unmarshal_LogConfig(t *testing.T) {
	testCases := []struct {
		name        string
		env         map[string]string
		overrides   map[string]interface{}
		cfgFile     string
		expectError string
		expect      func(*testConfig)
	}{
		{
			name:    "empty file should populate defaults",
			cfgFile: "testdata/empty/config.yaml",
		},
		{
			name:    "missing file should populate defaults",
			cfgFile: "testdata/does-not-exist/config.yaml",
		},
		{
			name:    "file should override defaults",
			cfgFile: "testdata/overrides/config.yaml",
			expect: func(c *testConfig) {
				c.Logging.Console.Level = "trace"
				c.Logging.File.Enabled = false
				c.Logging.File.KeepFiles = 100
			},
		},
		{
			name: "env vars should override defaults",
			env: map[string]string{
				"MYKEY_LOGGING_CONSOLE_LEVEL":  "error",
				"MYKEY_LOGGING_FILE_ENABLED":   "false",
				"MYKEY_LOGGING_FILE_KEEPFILES": "200",
			},
			expect: func(c *testConfig) {
				c.Logging.Console.Level = "error"
				c.Logging.File.Enabled = false
				c.Logging.File.KeepFiles = 200
			},
		},
		{
			name: "overrides should override defaults",
			overrides: map[string]interface{}{
				"mykey.logging.console.level":  "debug",
				"mykey.logging.file.enabled":   false,
				"mykey.logging.file.keepfiles": 300,
			},
			expect: func(c *testConfig) {
				c.Logging.Console.Level = "debug"
				c.Logging.File.Enabled = false
				c.Logging.File.KeepFiles = 300
			},
		},
		{
			name:    "environment should override config file",
			cfgFile: "testdata/overrides/config.yaml",
			env: map[string]string{
				"MYKEY_LOGGING_CONSOLE_LEVEL":  "error",
				"MYKEY_LOGGING_FILE_ENABLED":   "true",
				"MYKEY_LOGGING_FILE_KEEPFILES": "200",
			},
			expect: func(c *testConfig) {
				c.Logging.Console.Level = "error"
				c.Logging.File.Enabled = true
				c.Logging.File.KeepFiles = 200
			},
		},
		{
			name:    "overrides should override config file and environment",
			cfgFile: "testdata/overrides/config.yaml",
			env: map[string]string{
				"MYKEY_LOGGING_CONSOLE_LEVEL":  "error",
				"MYKEY_LOGGING_FILE_ENABLED":   "true",
				"MYKEY_LOGGING_FILE_KEEPFILES": "200",
			},
			overrides: map[string]interface{}{
				"mykey.logging.console.level":  "debug",
				"mykey.logging.file.enabled":   false,
				"mykey.logging.file.keepfiles": 300,
			},
			expect: func(c *testConfig) {
				c.Logging.Console.Level = "debug"
				c.Logging.File.Enabled = false
				c.Logging.File.KeepFiles = 300
			},
		},
		{
			name:        "validation failures should return a useful error",
			cfgFile:     "testdata/invalid/config.yaml",
			expectError: `(?s)invalid configuration under key "mykey", please check config file "testdata/invalid/config.yaml" and "THELMA_.*" environment variables:.*"mykey.logging.console.level" value this-is-not-a-valid-log-level does not match constraint "oneof: trace debug info warn error"`,
		},
		{
			name:    "nested maps and arrays should deserialize correctly",
			cfgFile: "testdata/nested/config.yaml",
			expect: func(c *testConfig) {
				c.Docker.HostAliases = []HostAlias{
					{
						Name: "foo",
						Addr: "127.0.0.1",
					}, {
						Name: "bar",
						Addr: "10.11.12.13",
					},
				}

				c.Airports.IataCodes = map[string]Airport{
					"BOS": {
						Name: "Logan",
						Location: Location{
							City:    "Boston",
							Country: "US",
						},
					},
					"LHR": {
						Name: "Heathrow",
						Location: Location{
							City:    "London",
							Country: "GB",
						},
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envPrefix := randEnvPrefix()
			configKey := "mykey"

			shadowEnv(t, envPrefix, tc.env)

			cfg, err := Load(func(options Options) Options {
				options.ConfigFile = tc.cfgFile
				options.EnvPrefix = envPrefix

				overrides := map[string]interface{}{
					"home": "fake/home/does/not/exist",
				}
				for k, v := range tc.overrides {
					overrides[k] = v
				}
				options.Overrides = overrides
				return options
			})

			if !assert.NoError(t, err) {
				return
			}

			expected := &testConfig{}
			defaults.SetDefaults(expected)
			if tc.expect != nil {
				tc.expect(expected)
			}

			actual := &testConfig{}

			err = cfg.Unmarshal(configKey, actual)

			if tc.expectError != "" {
				assert.Error(t, err)
				assert.Regexp(t, tc.expectError, err)
				return
			}

			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, expected, actual)
		})
	}
}

func TestConfig_Home(t *testing.T) {
	fakeHome := "relative/path/on/filesystem"

	cfg, err := Load(func(options Options) Options {
		// don't use default config file or env prefix, to make sure we don't pick up random values from the environment
		options.ConfigFile = "do/not/use/default/config/file"
		options.EnvPrefix = "do/not/use/default/prefix"
		options.Overrides = map[string]interface{}{
			"home": fakeHome,
		}
		return options
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.True(t, filepath.IsAbs(cfg.Home()))
	assert.True(t, strings.HasSuffix(cfg.Home(), fakeHome))
}

func TestConfig_Dump(t *testing.T) {
	expected := map[string]interface{}{
		"key1": 2,
		"key2": "foobar",
		"home": "does/not/exist",
	}

	cfg, err := Load(func(options Options) Options {
		// don't use default config file or env prefix, to make sure we don't pick up random values from the environment
		options.ConfigFile = "do/not/use/default/config/file"
		options.EnvPrefix = "do-not-use-default-prefix"
		options.Overrides = expected
		return options
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expected, cfg.Dump())
}

// return a random environment variable prefix for a single test case
func randEnvPrefix() string {
	return fmt.Sprintf("THELMA_%s_", testutils.RandString(6))
}

// override environment variables for the duration of a single test case.
func shadowEnv(t *testing.T, prefix string, vars map[string]string) {
	type envVar struct {
		originalValue string
		defined       bool
	}

	shadowed := make(map[string]envVar, len(vars))

	for name, newValue := range vars {
		fullName := fmt.Sprintf("%s%s", prefix, name)
		val, defined := os.LookupEnv(fullName)

		shadowed[fullName] = envVar{
			originalValue: val,
			defined:       defined,
		}
		if err := os.Setenv(fullName, newValue); err != nil {
			t.Fatalf("Unexpected error setting env var %s=%s: %v", fullName, newValue, err)
		}
	}

	t.Cleanup(func() {
		for n, v := range shadowed {
			if !v.defined {
				if err := os.Unsetenv(n); err != nil {
					t.Fatalf("Unexpected error unsetting env var %s: %v", n, err)
				}
				return
			}
			if err := os.Setenv(n, v.originalValue); err != nil {
				t.Fatalf("Unexpected error setting env var %s=%s: %v", n, v.originalValue, err)
			}
		}
	})
}
