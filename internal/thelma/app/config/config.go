package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/mcuadros/go-defaults"
	"github.com/rs/zerolog/log"
	"os"
	"reflect"
	"strings"
)

// delimiter used for configuration keys in koanf
const keyDelimiter = "."

// delimiter used for environment variables in koanf
const envDelimiter = "_"

// HomeKey configuration key used for thelma home (path to terra-helmfile clone)
const HomeKey = "home"

// Config is the configuration utility for Thelma. See README.md for usage examples.
type Config interface {
	// Unmarshal will unmarshal all configuration data under the given prefix into the target struct.
	//
	// Unmarshal supports use of annotations from the Validator (https://github.com/go-playground/validator)
	// and Defaults (https://github.com/mcuadros/go-defaults) libraries in structs.
	Unmarshal(prefix string, into interface{}) error
	// Dump returns all config values as a map for debugging purposes
	Dump() map[string]interface{}
	// Home returns a path the local terra-helmfile clone
	Home() string
}

type config struct {
	options   Options
	koanf     *koanf.Koanf
	validator *validator.Validate
}

// NewTestConfig is for use in unit tests.
// It creates a new config object, applying the given settings.
// By default it sets "home" to the OS temp dir, but this be overridden in the settings map.
// It DOES NOT include any configuration from the environment or config files (~/.thelma/config.yaml)
func NewTestConfig(settings map[string]interface{}) (Config, error) {
	// global defaults here.
	overrides := map[string]interface{}{
		HomeKey: os.TempDir(),
	}
	// merge user-supplied settings with defaults
	for k, v := range settings {
		overrides[k] = v
	}

	return Load(func(options Options) Options {
		options.ConfigFile = ""       // Disable config file loading
		options.EnvPrefix = ""        // Disable env var loading
		options.Overrides = overrides // Pass in user-supplied settings as overrides
		return options
	})
}

// Load Thelma configuration from file, environment, etc into a new Config
func Load(opts ...Option) (Config, error) {
	options := DefaultOptions()
	for _, option := range opts {
		options = option(options)
	}

	_koanf := koanf.New(keyDelimiter)

	// load config from file ~/.thelma/config.yaml
	if options.ConfigFile != "" {
		if _, err := os.Stat(options.ConfigFile); os.IsNotExist(err) {
			log.Warn().Msgf("No config file found at %s, won't try to load it", options.ConfigFile)
		} else if err != nil {
			return nil, fmt.Errorf("error checking if configuration file %s exists: %v", options.ConfigFile, err)
		} else {
			if err := _koanf.Load(file.Provider(options.ConfigFile), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("error loading configuration from file %s: %v", options.ConfigFile, err)
			}
		}
	}

	// load config from env vars (eg. THELMA_HOME)
	if options.EnvPrefix != "" {
		envProvider := env.Provider(options.EnvPrefix, keyDelimiter, envVarReplacer(options.EnvPrefix))
		if err := _koanf.Load(envProvider, nil); err != nil {
			return nil, fmt.Errorf("error reading configuration from environment: %v", err)
		}
	}

	// apply configuration overrides (these are used in tests)
	overrideProvider := confmap.Provider(options.Overrides, keyDelimiter)
	if err := _koanf.Load(overrideProvider, nil); err != nil {
		return nil, fmt.Errorf("error applying configuration overrides: %v", err)
	}

	// validate configuration
	if !_koanf.Exists(HomeKey) || _koanf.Get(HomeKey) == "" {
		return nil, fmt.Errorf("please specify path to terra-helmfile clone, via the THELMA_HOME environment variable or via the `home:` setting in %s", options.ConfigFile)
	}

	return &config{
		options:   options,
		koanf:     _koanf,
		validator: validator.New(),
	}, nil
}

func (c *config) Unmarshal(configPrefix string, into interface{}) error {
	// Make sure we were passed a struct pointer
	value := reflect.ValueOf(into)
	if value.Kind() != reflect.Ptr {
		panic(fmt.Errorf("expected struct pointer, got %v: %v", value.Kind(), into))
	}
	if value.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("expected struct pointer, got %v pointer: %v", value.Elem().Kind(), into))
	}

	// Set defaults and make sure they pass validation
	defaults.SetDefaults(into)
	if err := c.validator.Struct(into); err != nil {
		panic(fmt.Errorf("struct defaults do not pass validation: %v", err))
	}

	if err := c.koanf.Unmarshal(configPrefix, into); err != nil {
		return fmt.Errorf("error unmarshalling config key %s into struct: %v", configPrefix, err)
	}

	// Verify configuration passes validation constraints
	return c.validateStruct(into, configPrefix)
}

func (c *config) validateStruct(s interface{}, configPrefix string) error {
	err := c.validator.Struct(s)
	if err == nil {
		return nil
	}

	// the rest of this is just about generating user-friendly error messages

	// make a useful error message header suggesting where the potential bad config values might live
	errHeader := fmt.Sprintf(`invalid configuration under key %q, please check config file %q and %q environment variables:`, configPrefix, c.options.ConfigFile, c.options.EnvPrefix)

	// if we got an unexpected error back, return it as-is
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return fmt.Errorf("%s %v", errHeader, err)
	}

	// for some reason the validation library does not include the validation constraint in error messages,
	// which makes it hard to understand why a particular config value is being rejected. So we generate a better
	// message that includes the constraint here.

	// build a slice of all validation errors
	var msgs []string
	for _, verr := range validationErrors {
		// get the config key that caused the error, eg. "logging.console.level"
		configKey := fieldNameToConfigKey(verr, configPrefix)
		// get bad value, eg. "info-with-typo"
		value := verr.Value()
		// get constraint, eg."oneof: trace debug info warn error"
		constraint := fmt.Sprintf("%s: %s", verr.Tag(), verr.Param())

		// make a summary message
		msg := fmt.Sprintf("%q value %v does not match constraint %q", configKey, value, constraint)

		// log the underlying error at debug level, but don't include at info-level because it's usually just noise.
		log.Debug().Str("key", configKey).Str("constraint", constraint).Interface("value", value).Msgf("configuration error: %v", verr)

		// append to list of all errors
		msgs = append(msgs, fmt.Sprintf("  %s", msg))
	}
	return fmt.Errorf("%s\n%s", errHeader, strings.Join(msgs, "\n"))
}

// convert struct field name like "logConfig.File.Level" to configuration key like
// "logging.file.level", for use in validation error messages
func fieldNameToConfigKey(verr validator.FieldError, configPrefix string) string {
	// eg. logConfig.File.Level
	structPath := verr.StructNamespace()

	// downcase field names. eg.
	// "logConfig.File.Level" -> "logconfig.file.level"
	structPath = strings.ToLower(structPath)

	// replace struct type name with config prefix. eg.
	// "logConfig.file.level" -> "logging.file.level"
	tokens := strings.Split(structPath, ".")
	if len(tokens) > 0 {
		tokens[0] = configPrefix
	}

	return strings.Join(tokens, keyDelimiter)
}

func (c *config) Home() string {
	return c.koanf.String(HomeKey)
}

func (c *config) Dump() map[string]interface{} {
	return c.koanf.All()
}

func envVarReplacer(envPrefix string) func(string) string {
	return func(envVar string) string {
		configKey := strings.TrimPrefix(envVar, envPrefix)
		configKey = strings.ToLower(configKey)
		configKey = strings.ReplaceAll(configKey, envDelimiter, keyDelimiter)
		return configKey
	}
}
