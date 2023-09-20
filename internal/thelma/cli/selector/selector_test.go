package selector

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sort"
	"strings"
	"testing"
)

func Test_Selector(t *testing.T) {
	testCases := []struct {
		name           string
		args           string
		expectErr      string
		expectReleases []string
		setupFn        func() error
	}{
		{
			name:      "empty input",
			args:      "",
			expectErr: "please specify a target environment or cluster",
		},
		{
			name:      "one chart multiple releases",
			args:      "sam",
			expectErr: "please specify a target environment or cluster",
		},
		{
			name:      "multiple charts, no environment scoping",
			args:      "-r agora,sam,rawls",
			expectErr: "please specify a target environment or cluster",
		},
		{
			name:           "multiple charts, scope to dev",
			args:           "-r agora,sam,rawls -e dev",
			expectReleases: []string{"agora-dev", "sam-dev"},
		},
		{
			name:           "multiple charts, scope to staging",
			args:           "-r agora,sam,rawls -e staging",
			expectReleases: []string{"rawls-staging", "sam-staging"},
		},
		{
			name:           "env selector: dev",
			args:           "-e dev ALL",
			expectReleases: []string{"agora-dev", "sam-dev"},
		},
		{
			name:           "env selector: my-bee",
			args:           "-e my-bee ALL",
			expectReleases: []string{"cromwell-my-bee"},
		},
		{
			name:           "cluster selector",
			args:           "-c terra-dev ALL",
			expectReleases: []string{"secrets-manager-terra-dev", "yale-terra-dev"},
		},
		{
			name:           "exact releases",
			args:           "--exact-release rawls-staging,yale-terra-dev,agora-dev",
			expectReleases: []string{"agora-dev", "rawls-staging", "yale-terra-dev"},
		},
		{
			name:           "multiple environments",
			args:           "--environment dev,swatomation --release ALL",
			expectReleases: []string{"agora-dev", "sam-dev", "workspacemanager-swatomation"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFn != nil {
				err := tc.setupFn()
				require.NoError(t, err)
			}

			statefixture, err := statefixtures.LoadFixtureFromFile("testdata/statefixture.yaml")
			require.NoError(t, err)

			selector := NewSelector()

			var selection []terra.Release
			cobraCommand := &cobra.Command{RunE: func(cmd *cobra.Command, args []string) error {
				var cmdErr error
				selection, cmdErr = selector.GetSelection(statefixture.Mocks().State, cmd.Flags(), args)
				return cmdErr
			}}

			selector.AddFlags(cobraCommand)

			cobraCommand.SetArgs(strings.Fields(tc.args))

			err = cobraCommand.Execute()
			if tc.expectErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)

			names := releaseFullNames(selection).Elements()
			sort.Strings(names)
			assert.Equal(t, tc.expectReleases, names)
		})
	}
}
