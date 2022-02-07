# Logging

This package handles logging for Thelma.

It uses the [zerolog](https://github.com/rs/zerolog) and [lumberjack](https://github.com/natefinch/lumberjack) libraries, and supports logging to two locations:
* console (stderr): pretty-formatted, defaults to `info` level.
* `~/.thelma/logs/thelma.log`: JSON-formatted, defaults to `debug` level, [automatically rotated](https://github.com/natefinch/lumberjack)

## Configuration
Logging settings can be changed via Thelma configuration. Eg.
```
logging:
  console:
    level: debug
  file:
    level: trace
```
See the logConfig struct in `logging.go` for more configuration options.

### Including Caller Information in Logs

Caller information (source file and line number) can be included in logs by setting `logging.caller.enabled` or `THELMA_LOGGING_CALLER_ENABLED` to `true` in config file or environment, respectively.  

## Usage

Clients should log messages using `log.Logger`:

```
import "github.com/rs/zerolog/log"

func doSomething() {
  	logger.Info().Str("my-useful-field", "blah").Msgf("An interesting value: %d", 123)
}
```