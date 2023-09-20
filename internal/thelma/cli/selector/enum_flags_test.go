package selector

import (
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newReleasesFlag(t *testing.T) {
	t.Run("just flag", func(t *testing.T) {
		flag := newReleasesFlag()
		pflags := &pflag.FlagSet{}
		pflags.AddFlag(&pflag.Flag{
			Name:    flag.flagName,
			Changed: true,
		})
		out, err := flag.preProcessHook([]string{allSelector}, []string{}, pflags)
		assert.Equal(t, []string{allSelector}, out)
		assert.NoError(t, err)
	})
	t.Run("just arg", func(t *testing.T) {
		flag := newReleasesFlag()
		pflags := &pflag.FlagSet{}
		pflags.AddFlag(&pflag.Flag{
			Name:    flag.flagName,
			Changed: false,
		})
		out, err := flag.preProcessHook([]string{}, []string{allSelector}, pflags)
		assert.Equal(t, []string{allSelector}, out)
		assert.NoError(t, err)
	})
	t.Run("not both flag and arg", func(t *testing.T) {
		flag := newReleasesFlag()
		pflags := &pflag.FlagSet{}
		pflags.AddFlag(&pflag.Flag{
			Name:    flag.flagName,
			Changed: true,
		})
		_, err := flag.preProcessHook([]string{allSelector}, []string{allSelector}, pflags)
		assert.ErrorContains(t, err, "not both")
	})
	t.Run("disallow empty normally", func(t *testing.T) {
		flag := newReleasesFlag()
		pflags := &pflag.FlagSet{}
		pflags.AddFlag(&pflag.Flag{
			Name:    flag.flagName,
			Changed: false,
		})
		_, err := flag.preProcessHook([]string{}, []string{}, pflags)
		assert.ErrorContains(t, err, "at least one")
	})
	t.Run("allow empty if exact set", func(t *testing.T) {
		flag := newReleasesFlag()
		pflags := &pflag.FlagSet{}
		pflags.AddFlag(&pflag.Flag{
			Name:    flag.flagName,
			Changed: false,
		})
		pflags.AddFlag(&pflag.Flag{
			Name:    flagNames.exactRelease,
			Changed: true,
		})
		out, err := flag.preProcessHook([]string{}, []string{}, pflags)
		assert.Equal(t, []string{allSelector}, out)
		assert.NoError(t, err)
	})
}

func Test_newExactReleasesFlag(t *testing.T) {
	t.Run("empty = all", func(t *testing.T) {
		flag := newExactReleasesFlag()
		pflags := &pflag.FlagSet{} // not used by function
		out, err := flag.preProcessHook([]string{}, []string{}, pflags)
		assert.Equal(t, []string{allSelector}, out)
		assert.NoError(t, err)
	})
	t.Run("specifying all fails", func(t *testing.T) {
		flag := newExactReleasesFlag()
		pflags := &pflag.FlagSet{} // not used by function
		_, err := flag.preProcessHook([]string{allSelector}, []string{}, pflags)
		assert.ErrorContains(t, err, "cannot be used")
	})
	t.Run("flag values passed unprocessed", func(t *testing.T) {
		flag := newExactReleasesFlag()
		pflags := &pflag.FlagSet{} // not used by function
		flagValues := []string{"foo", "bar"}
		out, err := flag.preProcessHook(flagValues, []string{}, pflags)
		assert.Equal(t, flagValues, out)
		assert.NoError(t, err)
	})
}
