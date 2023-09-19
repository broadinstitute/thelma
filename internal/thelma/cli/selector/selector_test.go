package selector

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sort"
	"strings"
	"testing"
)

func Test_Selector(t *testing.T) {
	tmpdir := t.TempDir()

	testCases := []struct {
		name           string
		args           string
		expectErr      string
		expectReleases []string
		options        func(options *Options)
		setupFn        func() error
	}{
		{
			name:      "empty input",
			args:      "",
			expectErr: "please specify at least one release",
		},
		{
			name:           "one chart multiple releases",
			args:           "sam",
			expectReleases: []string{"sam-dev", "sam-staging"},
		},
		{
			name:           "multiple charts",
			args:           "-r agora,sam,rawls",
			expectReleases: []string{"agora-dev", "rawls-staging", "sam-dev", "sam-staging"},
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
			name: "changed files list",
			args: "--changed-files-list " + tmpdir + "/changedfiles.txt",
			setupFn: func() error {
				return os.WriteFile(tmpdir+"/changedfiles.txt", []byte(`charts/workspacemanager/somefile.txt`), 0644)
			},
			expectReleases: []string{"workspacemanager-swatomation"},
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
		{
			name:           "destination type: env",
			args:           "--destination-type environment ALL",
			expectReleases: []string{"agora-dev", "rawls-staging", "sam-dev", "sam-staging", "workspacemanager-swatomation"},
		},
		{
			name:           "destination type: cluster",
			args:           "--destination-type cluster ALL",
			expectReleases: []string{"secrets-manager-terra-dev", "yale-terra-dev", "yale-terra-staging"},
		},
		{
			name:           "destination base: bee (bees excluded by defaults)",
			args:           "--destination-base bee ALL",
			expectReleases: []string{"workspacemanager-swatomation"},
		},
		{
			name:           "destination base: live",
			args:           "--destination-base live ALL",
			expectReleases: []string{"agora-dev", "rawls-staging", "sam-dev", "sam-staging"},
		},
		{
			name:           "environment lifecycle: static, destination type: env",
			args:           "--environment-lifecycle=static --destination-type=environment ALL",
			expectReleases: []string{"agora-dev", "rawls-staging", "sam-dev", "sam-staging"},
		},
		{
			name:           "environment lifecycle: template, destination type: env",
			args:           "--environment-lifecycle=template --destination-type=environment ALL",
			expectReleases: []string{"workspacemanager-swatomation"},
		},
		{
			name:           "environment lifecycle: dynamic, destination type: env",
			args:           "--environment-lifecycle=dynamic --destination-type=environment ALL",
			expectReleases: []string{"cromwell-my-bee"},
		},
		{
			name:           "environment template: swatomation, environment lifecycle: dynamic, destination type: env",
			args:           "--environment-template=swatomation --environment-lifecycle=dynamic --destination-type=environment ALL",
			expectReleases: []string{"cromwell-my-bee"},
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

			var options []Option
			if tc.options != nil {
				options = append(options, tc.options)
			}

			selector := NewSelector(options...)

			var selection *Selection
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

			names := releaseFullNames(selection.Releases).Elements()
			sort.Strings(names)
			assert.Equal(t, tc.expectReleases, names)
		})
	}
}
