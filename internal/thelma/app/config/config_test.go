package config

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestConfigKeyNamesMatchYamlTags(t *testing.T) {
	dataType := reflect.TypeOf(data{})
	keysType := reflect.TypeOf(Keys)

	assert.Positive(t, dataType.NumField(), "config.data type should have at least one field")
	assert.Equal(t, dataType.NumField(), keysType.NumField(), "config.data and config.Keys should have same number of fields")

	// Iterate through the fields of the config data type and make sure there is a corresponding Key field
	for i := 0; i < dataType.NumField(); i++ {
		fieldName := dataType.Field(i).Name
		yamlTag := dataType.Field(i).Tag.Get("yaml")

		if yamlTag == "" {
			t.Fatalf("config.data field %s is missing a yaml tag", fieldName)
		}

		keyField := reflect.ValueOf(Keys).FieldByName(fieldName)
		if keyField.IsZero() {
			t.Fatalf("config.Keys should have field %s corresponding to config.data field %s, but it does not", fieldName, fieldName)
		}

		actual := keyField.String()
		assert.Equal(t, yamlTag, actual, "Expected Keys.%s to equal config.data.%s yaml tag %s, got %s", fieldName, fieldName, yamlTag, actual)
	}
}
