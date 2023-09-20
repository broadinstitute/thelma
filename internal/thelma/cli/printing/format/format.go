package format

import (
	"encoding/json"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
)

// IsSupported returns true if the given format is supported
func IsSupported(formatName string) bool {
	var f Format
	err := (&f).FromString(formatName)
	return err == nil
}

// SupportedFormats returns the names of supported formats as strings
func SupportedFormats() []string {
	var result []string
	for f := range formats {
		result = append(result, f.String())
	}
	return result
}

// Format is an enum type representing different output formats
type Format int

const (
	// Yaml format prints output in YAML
	Yaml Format = iota
	// Json format prints output in JSON
	Json
	// None format causes output to be suppressed
	None
	// PrettyYaml format prints output in colored YAML
	PrettyYaml
)

// Format will write formatted data to the given writer
func (f Format) Format(data interface{}, w io.Writer) error {
	fn := formats[f]
	return fn(data, w)
}

// FromString will set the receiver's value to the one denoted by the given string
func (f *Format) FromString(value string) error {
	switch value {
	case "yaml":
		*f = Yaml
		return nil
	case "pretty-yaml":
		*f = PrettyYaml
		return nil
	case "json":
		*f = Json
		return nil
	case "none":
		*f = None
		return nil
	}
	return errors.Errorf("unknown format: %q", value)
}

// String returns a string representation of this format
func (f Format) String() string {
	switch f {
	case Yaml:
		return "yaml"
	case Json:
		return "json"
	case None:
		return "none"
	case PrettyYaml:
		return "pretty-yaml"
	}
	return "unknown"
}

// Type alias for functions that write data in a given format to an io.Writer
type formatFn func(data interface{}, w io.Writer) (err error)

// Store output formats in a map, keyed by name
var formats = map[Format]formatFn{
	Yaml:       formatYaml,
	PrettyYaml: formatPrettyYaml,
	Json:       formatJson,
	None:       formatNone,
}

func formatJson(data interface{}, w io.Writer) error {
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent("", "  ")
	return jsonEncoder.Encode(data)
}

func formatYaml(data interface{}, w io.Writer) error {
	yamlEncoder := yaml.NewEncoder(w)
	yamlEncoder.SetIndent(2)
	return yamlEncoder.Encode(data)
}

func formatNone(_ interface{}, w io.Writer) error {
	return nil
}
