package credentials

import "github.com/broadinstitute/thelma/internal/thelma/app/logging"

// withMasking decorates TokenProvider by configuring Thelma's logger to mask any secrets it returns
func withMasking(p TokenProvider) TokenProvider {
	return &maskingProvider{
		inner: p,
	}
}

type maskingProvider struct {
	inner TokenProvider
}

func (m *maskingProvider) Get() ([]byte, error) {
	return m.mask(m.inner.Get())
}

func (m *maskingProvider) Reissue() ([]byte, error) {
	return m.mask(m.inner.Reissue())
}

func (m *maskingProvider) mask(secret []byte, err error) ([]byte, error) {
	if err != nil || len(secret) == 0 {
		return secret, err
	}

	logging.MaskSecret(string(secret))

	return secret, err
}
