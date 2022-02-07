# Config

This package handles configuration for Thelma.

It supports configuration through:
* a configuration file, located at `~/.thelma/config.yaml`
* environment variables (eg. `THELMA_HOME`)

It uses [Koanf](https://github.com/knadh/koanf) under the hood to provide these features.

### Usage

Declare a config struct. Then, unmarshal configuration into it at runtime using `config.Unmarshal`.

You can optionally:
* Set default values for fields using the `default` struct tag (via [defaults](https://github.com/mcuadros/go-defaults) library) 
* Add validation to config values using the `validate` struct tag (via [validator](https://github.com/go-playground/validator) library)

### Example
```
// Declare a config struct with optional default and validation tags
type LogConfig struct {
	Logging struct {
		Console struct {
			Level string `default:"info" validate:"oneof=trace debug info warn error"`
		}
		File struct {
			Enabled   bool   `default:"true"`
			Level     string `default:"debug" validate:"oneof=trace debug info warn error"`
			KeepFiles int    `description:"Number of rotated log files to keep" default:"5" validate:"gte=0"`
			MaxSizeMb int    `description:"Max size of rotated log files in megabytes" default:"8" validate:"gte=0"`
		}
	}
}

// Then use UnmarshalKey to unmarshal all config values under a given
// prefix into the struct.
func setupLogging(cfg config.Config) {
    var logConfig LogConfig
    if err := cfg.UnmarshalKey("logging", &logConfig) {
       // handle error
    }
    // initialize logging based on LogConfig settings
    logging.Level = logConfig.Level
    // ...
}
```

### Root Configuration Values

Most clients of config will bundle all configuration under a unique key (eg. `logging`); this config should be deserialized into a struct using `UnmarshalKey()`. However, a few core, system-wide configuration values exist at the root level.

These include:
* **`home`**: refers to the path where terra-helmfile is checked out

These config values are retrieved using methods on the Config object itself. (eg. `cfg.Home()`)

Only truly global, cross-cutting values that have implications for many parts of Thelma should be added at the root level.

### FAQ

#### Why Koanf and not [Viper](https://github.com/spf13/viper)?

Viper has a really unfortunate [long-running bug](https://github.com/spf13/viper/issues/761) where environment variables are ignored when unmarshalling config into a struct.

Although there are some hacky workarounds for the bug, they only work for Unmarshal(), and [not UnmarshalKey()](https://github.com/spf13/viper/issues/1012), because the two functions, for some reason, have very different implementations.

Koanf is not exposed as part of this package's API, so it should be relatively simple to migrate back to Viper if/when this issue is fixed. 
