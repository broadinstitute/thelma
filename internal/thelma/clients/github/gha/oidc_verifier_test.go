package gha

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_initOidcVerifier(t *testing.T) {
	thelmaConfig, err := config.Load(config.WithTestDefaults(t))
	require.NoError(t, err)

	var cfg ghaConfig
	err = thelmaConfig.Unmarshal(configPrefix, &cfg)
	require.NoError(t, err)

	err = initOidcVerifier(cfg.Oidc.ValidatingIssuer)
	require.NoError(t, err)
}

func Test_initOidcVerifier_reuse(t *testing.T) {
	verifierBefore := &oidc.IDTokenVerifier{}
	verifier = verifierBefore
	require.NoError(t, initOidcVerifier("some string"))
	require.Same(t, verifierBefore, verifier)
}
