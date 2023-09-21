package selector

import (
	"github.com/broadinstitute/thelma/internal/thelma/charts/changedfiles"
	"github.com/broadinstitute/thelma/internal/thelma/charts/source"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type changedFilesListFlag struct {
	changedFilesList string
}

func (c *changedFilesListFlag) addToCobraCommand(cobraCommand *cobra.Command) {
	cobraCommand.Flags().StringVar(
		&c.changedFilesList,
		flagNames.changedFilesList,
		"path/to/changed-files.txt",
		`Run for releases matching a newline-separated list of updated files in terra-helmfile`,
	)
}

func (c *changedFilesListFlag) processInput(f *filterBuilder, state terra.State, chartsDir source.ChartsDir, _ []string, pflags *pflag.FlagSet) error {
	if !pflags.Changed(flagNames.changedFilesList) {
		return nil
	}
	changedFiles := changedfiles.New(chartsDir, state)
	releaseFilter, err := changedFiles.ReleaseFilter(c.changedFilesList)
	if err != nil {
		return err
	}
	f.addReleaseFilter(releaseFilter)
	return nil
}

func newChangedFilesList() *changedFilesListFlag {
	return &changedFilesListFlag{}
}
