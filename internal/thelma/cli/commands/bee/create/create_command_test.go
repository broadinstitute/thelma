package create

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CreateHelp(t *testing.T) {
	_cli := cli.New(func(options *cli.Options) {
		options.AddCommand("bee", bee.NewBeeCommand())
		options.AddCommand("bee create", NewBeeCreateCommand())
		options.ConfigureThelma(func(thelmaBuilder builder.ThelmaBuilder) {
			thelmaBuilder.WithTestDefaults()
		})
		options.SetArgs([]string{"bee create", "--help"})
	})
	assert.NoError(t, _cli.Execute(), "--help should execute successfully")
}
