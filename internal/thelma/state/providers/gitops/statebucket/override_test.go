package statebucket

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Override_Versions(t *testing.T) {
	override := &Override{}
	empty := &Override{}

	assert.Equal(t, *empty, *override)

	override.SetAppVersion("1.2.3")
	assert.Equal(t, "1.2.3", override.AppVersion)
	override.UnsetAppVersion()
	assert.Equal(t, "", override.AppVersion)

	override.SetChartVersion("4.5.6")
	assert.Equal(t, "4.5.6", override.ChartVersion)
	override.UnsetChartVersion()
	assert.Equal(t, "", override.ChartVersion)

	override.SetTerraHelmfileRef("my-branch")
	assert.Equal(t, "my-branch", override.TerraHelmfileRef)
	override.UnsetTerraHelmfileRef()
	assert.Equal(t, "", override.TerraHelmfileRef)

	override.SetFirecloudDevelopRef("my-branch")
	assert.Equal(t, "my-branch", override.FirecloudDevelopRef)
	override.UnsetFirecloudDevelopRef()
	assert.Equal(t, "", override.FirecloudDevelopRef)

	assert.Equal(t, *empty, *override)
	override.SetAppVersion("1.2.3")
	override.SetChartVersion("4.5.6")
	override.SetTerraHelmfileRef("my-branch")
	override.SetFirecloudDevelopRef("my-branch")

	assert.NotEqual(t, *empty, *override)
	override.UnsetAll()
	assert.Equal(t, *empty, *override)
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
