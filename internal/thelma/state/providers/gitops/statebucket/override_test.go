package statebucket

import (
	"encoding/json"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Override_Versions(t *testing.T) {
	override := &Override{}
	empty := &Override{}

	assert.Equal(t, *empty, *override)

	override.PinVersions(terra.VersionOverride{AppVersion: "1.2.3"})
	assert.Equal(t, "1.2.3", override.Versions.AppVersion)

	override.PinVersions(terra.VersionOverride{ChartVersion: "4.5.6"})
	assert.Equal(t, "1.2.3", override.Versions.AppVersion)
	assert.Equal(t, "4.5.6", override.Versions.ChartVersion)

	override.PinVersions(terra.VersionOverride{TerraHelmfileRef: "pr-1"})
	assert.Equal(t, "1.2.3", override.Versions.AppVersion)
	assert.Equal(t, "4.5.6", override.Versions.ChartVersion)
	assert.Equal(t, "pr-1", override.Versions.TerraHelmfileRef)

	override.PinVersions(terra.VersionOverride{FirecloudDevelopRef: "pr-2"})
	assert.Equal(t, "1.2.3", override.Versions.AppVersion)
	assert.Equal(t, "4.5.6", override.Versions.ChartVersion)
	assert.Equal(t, "pr-1", override.Versions.TerraHelmfileRef)
	assert.Equal(t, "pr-2", override.Versions.FirecloudDevelopRef)

	override.UnpinVersions()
	assert.Equal(t, "", override.Versions.AppVersion)
	assert.Equal(t, "", override.Versions.ChartVersion)
	assert.Equal(t, "", override.Versions.TerraHelmfileRef)
	assert.Equal(t, "", override.Versions.FirecloudDevelopRef)
}

func Test_Override_EnableDisable(t *testing.T) {
	override := &Override{}

	assert.False(t, override.HasEnableOverride())
	assert.False(t, override.IsEnabled())

	override.Enable()
	assert.True(t, override.HasEnableOverride())
	assert.True(t, override.IsEnabled())

	override.Disable()
	assert.True(t, override.HasEnableOverride())
	assert.False(t, override.IsEnabled())
}

func Test_Override_MarshalAndUnmarshal(t *testing.T) {
	input := &Override{}
	output := &Override{}

	assert.False(t, input.HasEnableOverride())

	data, err := json.Marshal(input)
	require.NoError(t, err)
	err = json.Unmarshal(data, output)
	require.NoError(t, err)

	assert.False(t, output.HasEnableOverride())

	input.Enable()
	assert.True(t, input.HasEnableOverride())
	assert.True(t, input.IsEnabled())

	data, err = json.Marshal(input)
	require.NoError(t, err)
	err = json.Unmarshal(data, output)
	require.NoError(t, err)

	assert.True(t, output.HasEnableOverride())
	assert.True(t, output.IsEnabled())

	input.Disable()
	assert.True(t, input.HasEnableOverride())
	assert.False(t, input.IsEnabled())

	data, err = json.Marshal(input)
	require.NoError(t, err)
	err = json.Unmarshal(data, output)
	require.NoError(t, err)

	assert.True(t, output.HasEnableOverride())
	assert.False(t, output.IsEnabled())
}
