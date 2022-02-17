package bee

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BeeHelp(t *testing.T) {
	_cli := cli.New(func(options *cli.Options) {
		options.AddCommand("bee", NewBeeCommand())
		options.ConfigureThelma(func(thelmaBuilder builder.ThelmaBuilder) {
			thelmaBuilder.WithTestDefaults()
		})
		options.SetArgs([]string{"bee", "--help"})
	})
	assert.NoError(t, _cli.Execute(), "--help should execute successfully")
}
