package provision

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ProvisionHelp(t *testing.T) {
	_cli := cli.New(func(options *cli.Options) {
		options.AddCommand("bee", bee.NewSlackCommand())
		options.AddCommand("bee create", NewBeeProvisionCommand())
		options.ConfigureThelma(func(thelmaBuilder builder.ThelmaBuilder) {
			thelmaBuilder.WithTestDefaults(t)
		})
		options.SetArgs([]string{"bee", "provision", "--help"})
	})
	assert.NoError(t, _cli.Execute(), "--help should execute successfully")
}
