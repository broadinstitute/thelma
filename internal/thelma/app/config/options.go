package config

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/env"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"path"
	"testing"
)

// defaultConfigFile is the name of the default thelma config file.
const defaultConfigFile = "config.yaml"

// Option can be used to customize default Options struct
type Option func(*Options)

// Options configure the behavior of config.Load(). This interface is provided to support testing only and shouldn't be used
// during regular program execution.
type Options struct {
	// Overrides overrides to apply to the configuration (useful for testing)
	Overrides map[string]interface{}
	// ConfigFile which file to load configuration file from (default ~/.thelma/config.yaml)
	// Set to "" to skip loading configuration from a file.
	ConfigFile string
	// EnvPrefix which prefix to use when loading configuration from environment variables (default "THELMA")
	// Set to "" to skip loading configuration from environment variables.
	EnvPrefix string
	// Profile load a specific configuration profile instead of the one specified by THELMA_CONFIG_PROFILE
	Profile string
}

// DefaultOptions returns default options for Config.Load()
func DefaultOptions() Options {
	return Options{
		Overrides:  make(map[string]interface{}),
		ConfigFile: path.Join(root.Default().Dir(), defaultConfigFile),
		EnvPrefix:  env.EnvPrefix,
		Profile:    "",
	}
}

func WithThelmaRoot(thelmaRoot root.Root) Option {
	return func(options *Options) {
		options.ConfigFile = path.Join(thelmaRoot.Dir(), defaultConfigFile)
	}
}

// WithTestDefaults sets useful test defaults, including:
// * disabling config loading from config file and environment variables
// * pointing THELMA_HOME at a temporary directory
func WithTestDefaults(t *testing.T) Option {
	return func(options *Options) {
		options.ConfigFile = ""                  // Disable config file loading
		options.EnvPrefix = ""                   // Disable env var loading
		options.Overrides[HomeKey] = t.TempDir() // Set HomeKey to a tmp dir
		options.Profile = defaultProfile         // Make sure we don't accidentally select `ci` profile
	}
}

// WithOverrides merges overrides on top of any that have already been set
func WithOverrides(overrides map[string]interface{}) Option {
	return func(options *Options) {
		for k, v := range overrides {
			options.Overrides[k] = v
		}
	}
}

// WithOverride merges a single override on top of any that have already been set
func WithOverride(key string, value interface{}) Option {
	return func(options *Options) {
		options.Overrides[key] = value
	}
}
