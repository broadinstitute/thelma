package config

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"path"
)

// defaultEnvPrefix prefix to use for configuration environment variable overrides.
const defaultEnvPrefix = "THELMA_"

// defaultConfigFile is the name of the default thelma config file.
const defaultConfigFile = "config.yaml"

// Option can be used to customize default Options struct
type Option func(Options) Options

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
}

// DefaultOptions returns default options for Config.Load()
func DefaultOptions() Options {
	return Options{
		Overrides:  make(map[string]interface{}),
		ConfigFile: path.Join(root.Default().Dir(), defaultConfigFile),
		EnvPrefix:  defaultEnvPrefix,
	}
}

func WithThelmaRoot(thelmaRoot root.Root) Option {
	return func(options Options) Options {
		options.ConfigFile = path.Join(thelmaRoot.Dir(), defaultConfigFile)
		return options
	}
}
