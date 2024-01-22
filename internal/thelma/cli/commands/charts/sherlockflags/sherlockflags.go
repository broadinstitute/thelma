package sherlockflags

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/charts/releaser"
	"github.com/broadinstitute/thelma/internal/thelma/clients/sherlock"
	"github.com/spf13/cobra"
)

const sherlockProdURL = "https://sherlock.dsp-devops.broadinstitute.org"
const sherlockDevURL = "https://sherlock-dev.dsp-devops.broadinstitute.org"

type flagValues struct {
	description      string
	sherlock         []string
	softFailSherlock []string
}

var flagNames = struct {
	description      string
	sherlock         string
	softFailSherlock string
}{
	description:      "description",
	sherlock:         "sherlock",
	softFailSherlock: "soft-fail-sherlock",
}

type sherlockUpdaterFlags struct {
	flagVals flagValues
}

func (f *sherlockUpdaterFlags) Description() string {
	return f.flagVals.description
}

// SherlockUpdaterFlags adds sherlock update flags to a cobra command and supports converting those flags to a releaser.DeployedVersionUpdater
type SherlockUpdaterFlags interface {
	// AddFlags add  flags to a command
	AddFlags(*cobra.Command)
	// GetDeployedVersionUpdater should be called during a Run function to get a releaser.DeployedVersionUpdater that matches the given flags
	GetDeployedVersionUpdater(thelmaApp app.ThelmaApp, dryRun bool) (*releaser.DeployedVersionUpdater, error)
	// Description returns value of --description flag
	Description() string
}

// NewSherlockUpdaterFlags returns a new SherlockUpdaterFlags
func NewSherlockUpdaterFlags() SherlockUpdaterFlags {
	return &sherlockUpdaterFlags{}
}

func (f *sherlockUpdaterFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&f.flagVals.sherlock, flagNames.sherlock, []string{sherlockProdURL}, "Sherlock servers to use as versioning systems to release to")
	cmd.Flags().StringSliceVar(&f.flagVals.softFailSherlock, flagNames.softFailSherlock, []string{sherlockDevURL}, "Sherlock server to use as versioning systems to release to, always using soft-fail behavior")
	cmd.Flags().StringVarP(&f.flagVals.description, flagNames.description, "d", "", "The description to use for these version bumps on any Sherlock versioning systems")
}

func (f *sherlockUpdaterFlags) GetDeployedVersionUpdater(app app.ThelmaApp, dryRun bool) (*releaser.DeployedVersionUpdater, error) {
	var updater releaser.DeployedVersionUpdater

	// If we're dry-running, the updater will be empty so we don't mutate anything.
	if dryRun {
		return &updater, nil
	}
	if len(f.flagVals.sherlock) > 0 || len(f.flagVals.softFailSherlock) > 0 {
		for _, sherlockURL := range f.flagVals.sherlock {
			if sherlockURL != "" {
				client, err := app.Clients().Sherlock(func(options *sherlock.Options) {
					options.Addr = sherlockURL
				})
				if err != nil {
					return &updater, err
				}
				updater.SherlockUpdaters = append(updater.SherlockUpdaters, client)
			}
		}
		for _, sherlockURL := range f.flagVals.softFailSherlock {
			if sherlockURL != "" {
				client, err := app.Clients().Sherlock(func(options *sherlock.Options) {
					options.Addr = sherlockURL
				})
				if err != nil {
					return &updater, err
				}
				updater.SoftFailSherlockUpdaters = append(updater.SoftFailSherlockUpdaters, client)
			}
		}
	}

	return &updater, nil
}
