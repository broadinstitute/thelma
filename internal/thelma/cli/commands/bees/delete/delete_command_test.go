package delete

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bees"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BeesDeleteHelp(t *testing.T) {
	_cli := cli.New(func(options *cli.Options) {
		options.AddCommand("bees", bees.NewBeesCommand())
		options.AddCommand("bees delete", NewBeesDeleteCommand())
		options.ConfigureThelma(func(thelmaBuilder builder.ThelmaBuilder) {
			thelmaBuilder.WithTestDefaults(t)
		})
		options.SetArgs([]string{"bees", "delete", "--help"})
	})
	assert.NoError(t, _cli.Execute(), "--help should execute successfully")
}
