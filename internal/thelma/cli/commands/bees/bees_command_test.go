package bees

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_BeesHelp(t *testing.T) {
	_cli := cli.New(func(options *cli.Options) {
		options.AddCommand("bees", NewBeesCommand())
		options.ConfigureThelma(func(thelmaBuilder builder.ThelmaBuilder) {
			thelmaBuilder.WithTestDefaults(t)
		})
		options.SetArgs([]string{"bees", "--help"})
	})
	assert.NoError(t, _cli.Execute(), "--help should execute successfully")
}
