package gha

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ghaOidcProviderWithBehavior(t *testing.T) {
	for name, fn := range map[string]func(*ghaOidcProviderWithBehavior) ([]byte, error){
		"get":     func(p *ghaOidcProviderWithBehavior) ([]byte, error) { return p.Get() },
		"reissue": func(p *ghaOidcProviderWithBehavior) ([]byte, error) { return p.Reissue() },
	} {
		t.Run(name, func(t *testing.T) {
			t.Run("always", func(t *testing.T) {
				t.Run("value", func(t *testing.T) {
					provider := &ghaOidcProviderWithBehavior{
						behavior: "always",
						delegate: &credentials.MockTokenProvider{
							ReturnString: "foo",
						},
					}
					token, err := fn(provider)
					require.NoError(t, err)
					require.Equal(t, "foo", string(token))
				})
				t.Run("error", func(t *testing.T) {
					provider := &ghaOidcProviderWithBehavior{
						behavior: "always",
						delegate: &credentials.MockTokenProvider{
							ReturnErr: true,
						},
					}
					token, err := fn(provider)
					require.Error(t, err)
					require.Empty(t, token)
				})
			})
			t.Run("never", func(t *testing.T) {
				provider := &ghaOidcProviderWithBehavior{
					behavior: "never",
					delegate: &credentials.MockTokenProvider{
						FailIfCalled: t,
					},
				}
				token, err := fn(provider)
				require.NoError(t, err)
				require.Empty(t, token)
			})
			t.Run("opportunistic", func(t *testing.T) {
				t.Run("value", func(t *testing.T) {
					provider := &ghaOidcProviderWithBehavior{
						behavior: "opportunistic",
						delegate: &credentials.MockTokenProvider{
							ReturnString: "foo",
						},
					}
					token, err := fn(provider)
					require.NoError(t, err)
					require.Equal(t, "foo", string(token))
				})
				t.Run("error", func(t *testing.T) {
					provider := &ghaOidcProviderWithBehavior{
						behavior: "opportunistic",
						delegate: &credentials.MockTokenProvider{
							ReturnErr: true,
						},
					}
					token, err := fn(provider)
					require.NoError(t, err)
					require.Empty(t, token)
				})
			})
			t.Run("unknown", func(t *testing.T) {
				provider := &ghaOidcProviderWithBehavior{
					behavior: "blah",
					delegate: &credentials.MockTokenProvider{
						FailIfCalled: t,
					},
				}
				token, err := fn(provider)
				require.ErrorContains(t, err, "unknown *gha.ghaOidcProviderWithBehavior behavior: blah")
				require.Empty(t, token)
			})
		})
	}
}
