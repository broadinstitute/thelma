package pinflags

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	climocks "github.com/broadinstitute/thelma/internal/thelma/cli/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

func Test_LoadFromEnv(t *testing.T) {
	// build thelma app set up for testing
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)

	app, err := builder.NewBuilder().
		WithTestDefaults(t).
		UseCustomStateLoader(fixture.Mocks().StateLoader).
		Build()
	require.NoError(t, err)

	// set up an empty cobra command and mock run context
	cmd := &cobra.Command{}
	rc := &climocks.RunContext{}
	rc.EXPECT().CobraCommand().Return(cmd)

	// add the cli flags we want to test
	cmd.SetArgs([]string{"--from-env", "dev"})

	flags := NewPinFlags()
	flags.AddFlags(cmd)

	// execute the command to make Cobra parse cli flags
	require.NoError(t, cmd.Execute())

	// now, load the pinoptions and make sure they match what we expect
	pinOpts, err := flags.GetPinOptions(app, rc)
	require.NoError(t, err)

	// make sure global flags were not set
	assert.Equal(t, "", pinOpts.Flags.TerraHelmfileRef)

	// make sure version overrides were copied from dev environment
	assert.Equal(t, 8, len(pinOpts.FileOverrides))
	assert.Equal(t, "2d309b1645a0", pinOpts.FileOverrides["sam"].AppVersion)
	assert.Equal(t, "0.34.0", pinOpts.FileOverrides["sam"].ChartVersion)
	assert.Equal(t, "", pinOpts.FileOverrides["sam"].TerraHelmfileRef)
}

func Test_LoadFromFile(t *testing.T) {
	// write fake verisons.json file to temp dir
	file := path.Join(t.TempDir(), "versions.json")
	require.NoError(t, os.WriteFile(file, []byte(`
	{
		"sam": {
			"appVersion": "1.2.3",
			"chartVersion": "4.5.6",
			"terraHelmfileRef": "pr-1"
		}
	}
	`), 0600))

	// build thelma app set up for testing
	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(statefixtures.Default)
	require.NoError(t, err)
	app, err := builder.NewBuilder().
		WithTestDefaults(t).
		UseCustomStateLoader(fixture.Mocks().StateLoader).
		Build()
	require.NoError(t, err)
	require.NoError(t, err)

	// set up an empty cobra command and run context
	cmd := &cobra.Command{}
	rc := &climocks.RunContext{}
	rc.On("CobraCommand").Return(cmd)

	// set the flags we actually want to test
	cmd.SetArgs([]string{"--versions-file", file, "--versions-format", "json"})

	flags := NewPinFlags()
	flags.AddFlags(cmd)

	// execute the command to make Cobra parse cli flags
	require.NoError(t, cmd.Execute())

	// now, load the pinoptions and make sure they match what we expect
	pinOpts, err := flags.GetPinOptions(app, rc)
	require.NoError(t, err)

	// make sure global flags were not set
	assert.Equal(t, "", pinOpts.Flags.TerraHelmfileRef)

	// make sure version overrides were copied from file
	assert.Equal(t, 1, len(pinOpts.FileOverrides))
	assert.Equal(t, "1.2.3", pinOpts.FileOverrides["sam"].AppVersion)
	assert.Equal(t, "4.5.6", pinOpts.FileOverrides["sam"].ChartVersion)
	assert.Equal(t, "pr-1", pinOpts.FileOverrides["sam"].TerraHelmfileRef)
}

func Test_GlobalFlags(t *testing.T) {
	// build thelma app set up for testing
	app, err := builder.NewBuilder().WithTestDefaults(t).Build()
	require.NoError(t, err)

	// set up an empty cobra command and run context
	cmd := &cobra.Command{}
	rc := &climocks.RunContext{}
	rc.On("CobraCommand").Return(cmd)

	// set the flags we actually want to test
	cmd.SetArgs([]string{"--terra-helmfile-ref", "foo"})

	flags := NewPinFlags()
	flags.AddFlags(cmd)

	// execute the command to make Cobra parse cli flags
	require.NoError(t, cmd.Execute())

	// now, load the pinoptions and make sure they match what we expect
	pinOpts, err := flags.GetPinOptions(app, rc)
	require.NoError(t, err)

	// make sure global flags were not set
	assert.Equal(t, "foo", pinOpts.Flags.TerraHelmfileRef)

	// no version overrides should be set
	assert.Empty(t, pinOpts.FileOverrides)
}
