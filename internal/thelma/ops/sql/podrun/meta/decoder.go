package meta

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"k8s.io/utils/strings/slices"
	"reflect"
	"strings"
)

// Decoder can decode a simple struct into a map[string]string for use of Kubernetes annotations and labels.
// It also supports decoding a map[string]string back into a struct.
// Note that string fields are added to the map as-is; all other field types are encoded/decoded into a JSON string.
//
// Structs with pointer fields are not supported. This has been tested for numeric, bool, and string
// datatypes; other data types might not work. (For example, runes definitely do not work because they are
// not supported by mapstructure).
type Decoder[T any] struct {
	// Prefix optional prefix to add to keys when serializing. eg. "whatever.terra.bio/"
	Prefix string
	// TagName struct tag (default "key")
	TagName string
}

const defaultTagName = "k8smeta"

func (d Decoder[T]) MergeTo(s T, m map[string]string) error {
	m2, err := d.ToMap(s)
	if err != nil {
		return err
	}
	for k, v := range m2 {
		m[k] = v
	}
	return nil
}

func (d Decoder[T]) FromMap(m map[string]string) (*T, error) {
	withoutPrefix := make(map[string]string)
	for k, v := range m {
		if !strings.HasPrefix(k, d.Prefix) {
			continue
		}
		k2 := strings.TrimPrefix(k, d.Prefix)
		withoutPrefix[k2] = v
	}

	var a T

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: UnmarshalNonStringFieldsFromJSON(),
		TagName:    d.tagName(),
		Result:     &a,
	})
	if err != nil {
		return nil, fmt.Errorf("error decoding annotations: %v", err)
	}

	if err = decoder.Decode(withoutPrefix); err != nil {
		return nil, fmt.Errorf("error decoding annotations: %v", err)
	}
	return &a, nil
}

func (d Decoder[T]) ToMap(s T) (map[string]string, error) {
	m := make(map[string]string)

	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("require struct, got %T", s))
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		val := v.Field(i)

		name := field.Name
		tagval := field.Tag.Get(d.tagName())

		key := name

		tagVals := strings.Split(tagval, ",")
		if len(tagVals) > 0 && len(tagVals[0]) != 0 {
			key = tagVals[0]
		}

		omitEmpty := slices.Contains(tagVals, "omitempty")

		var s string
		if val.Kind() == reflect.String {
			s = val.String()
		} else {
			data, err := json.Marshal(val.Interface())
			if err != nil {
				return nil, fmt.Errorf("error marshalling to JSON: %v", err)
			}
			s = string(data)
		}

		if len(s) == 0 && omitEmpty {
			continue
		}

		m[d.Prefix+key] = s
	}

	return m, nil
}

func (d Decoder[T]) tagName() string {
	if d.TagName != "" {
		return d.TagName
	}
	return defaultTagName
}

func UnmarshalNonStringFieldsFromJSON() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			// we are at the top node instead of a leaf (map/struct)
			return data, nil
		}

		if t.Kind() == reflect.String {
			// we don't marshal strings, they're already in the format we need
			return data, nil
		}

		s, ok := data.(string)
		if !ok {
			return nil, fmt.Errorf("expected data to be a string, got %T", data)
		}

		v := reflect.New(t)
		p := v.Interface()
		if err := json.Unmarshal([]byte(s), p); err != nil {
			return nil, fmt.Errorf("error unmarshalling data: %v", err)
		}

		return p, nil
	}
}
