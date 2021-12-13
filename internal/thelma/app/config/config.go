package config

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// ConfigRepoName name to use to refer to terra-helmfile config repo in log & error messages
const ConfigRepoName = "terra-helmfile"

// envPrefix prefix to use for configuration environment variable overrides.
const envPrefix = "THELMA"

// defaultLogLevel default level for logging, valid options are whatever zerolog accepts (eg. "debug", "trace")
const defaultLogLevel = "info"

// Keys makes yaml serialization keys for Data fields available without reflection.
var Keys = struct {
	Home     string
	Tmpdir   string
	LogLevel string
}{
	Home:     "home",
	Tmpdir:   "tmpdir",
	LogLevel: "loglevel",
}

// Data is a mutable, serializable struct for building an immutable Config
type Data struct {
	Home     string `yaml:"home"`
	Tmpdir   string `yaml:"tmpdir"`
	LogLevel string `yaml:"loglevel"`
}

// Config represents global config for Thelma
type Config struct {
	data Data
}

func Load(overrides map[string]interface{}) (*Config, error) {
	_viper := viper.New()

	// Set defaults
	_viper.SetDefault(Keys.Home, "")
	_viper.SetDefault(Keys.LogLevel, defaultLogLevel)
	_viper.SetDefault(Keys.Tmpdir, os.TempDir())

	// Configure Viper:
	// automatically interpret env vars prefixed with THELMA_ as config settings
	_viper.SetEnvPrefix(envPrefix)
	// map dashes to underscores ("THELMA_LOG_LEVEL" is mapped to the key "log-level")
	_viper.SetEnvKeyReplacer(configKeyReplacer())
	// automatically load config values from environment
	_viper.AutomaticEnv()

	// apply configuration overrides (these are used in tests)
	for k, v := range overrides {
		_viper.Set(k, v)
	}

	// Validation
	// Make sure home dir is configured and exists
	homePath := _viper.GetString(Keys.Home)
	if homePath == "" {
		return nil, fmt.Errorf("please specify path to %s clone via the environment variable %s", ConfigRepoName, configKeyToEnvVar(Keys.Home))
	}
	fullPath, err := utils.ExpandAndVerifyExists(homePath, fmt.Sprintf("%s clone", ConfigRepoName))
	if err != nil {
		return nil, err
	}
	_viper.Set(Keys.Home, fullPath)

	// Make sure log level is valid
	logLevel := _viper.GetString(Keys.LogLevel)
	if _, err := zerolog.ParseLevel(logLevel); err != nil {
		log.Warn().Msgf("Invalid log level %v, setting to %s", logLevel, defaultLogLevel)
		_viper.Set(Keys.LogLevel, defaultLogLevel)
	}

	// Convert viper config to a simple immutable config struct and return
	data := Data{}
	if err := _viper.Unmarshal(&data); err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
	}

	return &Config{data: data}, nil
}

// Home is the path to a terra-helmfile clone
func (cfg *Config) Home() string {
	return cfg.data.Home
}

// LogLevel is the level at which Thelma should log
func (cfg *Config) LogLevel() string {
	return cfg.data.LogLevel
}

// Tmpdir directory where Thelma should create temporary files
func (cfg *Config) Tmpdir() string {
	return cfg.data.Tmpdir
}

// configKeyToEnvVar transform config key to corresponding environment variable
// eg.
// "log-level" -> "THELMA_LOG_LEVEL"
func configKeyToEnvVar(key string) string {
	prefix := strings.ToUpper(envPrefix)
	key = configKeyReplacer().Replace(key)
	key = strings.ToUpper(key)
	return fmt.Sprintf("%s_%s", prefix, key)
}

// configKeyReplacer returns a string replacer that substitutes "-" with "_"
func configKeyReplacer() *strings.Replacer {
	return strings.NewReplacer("-", "_")
}
