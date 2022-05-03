package argocd

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ArgoCDHelp(t *testing.T) {
	_cli := cli.New(func(options *cli.Options) {
		options.AddCommand("argocd", NewArgoCDCommand())
		options.ConfigureThelma(func(thelmaBuilder builder.ThelmaBuilder) {
			thelmaBuilder.WithTestDefaults(t)
		})
		options.SetArgs([]string{"argocd", "--help"})
	})
	assert.NoError(t, _cli.Execute(), "--help should execute successfully")
}
