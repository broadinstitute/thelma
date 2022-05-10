package statebucket

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_DynamicEnvironment_JSON_Marshaller(t *testing.T) {
	var d DynamicEnvironment

	assert.Nil(t, d.Overrides)

	expected := `{"name":"","template":"","overrides":{},"hybrid":false,"fiab":{"ip":"","name":""}}`

	data, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))

	// make sure it works for pointer too
	data, err = json.Marshal(&d)
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))
}

func Test_DynamicEnvironment_JSON_Unmarshaller(t *testing.T) {
	var d DynamicEnvironment

	input := `{"name":"","template":"","overrides":null,"hybrid":false,"fiab":{"ip":"","name":""}}`
	err := json.Unmarshal([]byte(input), &d)
	require.NoError(t, err)

	assert.NotNil(t, d.Overrides)
}
