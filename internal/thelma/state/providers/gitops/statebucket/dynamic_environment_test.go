package statebucket

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func Test_DynamicEnvironment_JSON_Marshaller(t *testing.T) {
	var testsCases = []struct {
		name       string
		inputData  func(d *DynamicEnvironment)
		outputJSON string
	}{
		{
			name: "empty",
			inputData: func(d *DynamicEnvironment) {
				d.Overrides = make(map[string]*Override)
			},
			outputJSON: `
{
  "name": "",
  "template": "",
  "overrides": {},
  "hybrid": false,
  "fiab": {
    "ip": "",
    "name": ""
  },
  "terraHelmfileRef": "",
  "buildNumber": 0
}`,
		},
		{
			name: "with data",
			inputData: func(d *DynamicEnvironment) {
				d.Name = "e1"
				d.Template = "t1"
				d.TerraHelmfileRef = "deadbeef"
				d.BuildNumber = 12
				d.Hybrid = true
				d.Fiab.Name = "fiab1"
				d.Fiab.IP = "0.0.0.0"

				var override Override
				enabled := false
				override.Enabled = &enabled
				override.Versions.AppVersion = "1.2.3"
				override.Versions.ChartVersion = "4.5.6"
				override.Versions.FirecloudDevelopRef = "meh"
				override.Versions.TerraHelmfileRef = "r2"
				d.Overrides = map[string]*Override{
					"foo": &override,
				}
			},
			outputJSON: `
{
  "name": "e1",
  "template": "t1",
  "overrides": {
    "foo": {
      "enabled": false,
      "versions": {
        "appVersion": "1.2.3",
        "chartVersion": "4.5.6",
        "terraHelmfileRef": "r2",
        "firecloudDevelopRef": "meh"
      }
    }
  },
  "hybrid": true,
  "fiab": {
    "ip": "0.0.0.0",
    "name": "fiab1"
  },
  "terraHelmfileRef": "deadbeef",
  "buildNumber": 12
}`,
		},
	}

	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			var d DynamicEnvironment

			if tc.inputData != nil {
				tc.inputData(&d)
			}

			outputJSON := strings.TrimSpace(tc.outputJSON)

			data, err := json.MarshalIndent(d, "", "  ")
			require.NoError(t, err)
			assert.Equal(t, outputJSON, string(data))

			// make sure it works for pointer too
			data, err = json.MarshalIndent(&d, "", "  ")
			require.NoError(t, err)
			assert.Equal(t, outputJSON, string(data))

			// make sure it works in the other direction
			var d2 DynamicEnvironment
			err = json.Unmarshal([]byte(outputJSON), &d2)
			require.NoError(t, err)
			assert.Equal(t, d, d2)
		})
	}
}

func Test_DynamicEnvironment_JSON_Marshaller_ReplacesNilOverrides(t *testing.T) {
	var d DynamicEnvironment

	assert.Nil(t, d.Overrides)

	expected := `{"name":"","template":"","overrides":{},"hybrid":false,"fiab":{"ip":"","name":""},"terraHelmfileRef":"","buildNumber":0}`

	data, err := json.Marshal(d)
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))

	// make sure it works for pointer too
	data, err = json.Marshal(&d)
	require.NoError(t, err)
	assert.Equal(t, expected, string(data))
}

func Test_DynamicEnvironment_JSON_Unmarshaller_ReplacesNilOverrides(t *testing.T) {
	var d DynamicEnvironment

	input := `{"name":"","template":"","overrides":null,"hybrid":false,"fiab":{"ip":"","name":""},"terraHelmfileRef":"","buildNumber":0}`
	err := json.Unmarshal([]byte(input), &d)
	require.NoError(t, err)

	assert.NotNil(t, d.Overrides, "overrides should not be null even if it is in the JSON")
}
