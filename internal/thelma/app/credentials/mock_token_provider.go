package credentials

import (
	"github.com/pkg/errors"
	"testing"
)

// MockTokenProvider offers a TokenProvider that is as simple as possible
// to facilitate testing outside of this package.
type MockTokenProvider struct {
	ReturnString string
	ReturnBytes  []byte
	ReturnNil    bool
	ReturnErr    bool
	FailIfCalled *testing.T
}

func (m *MockTokenProvider) Get() ([]byte, error) {
	if m.FailIfCalled != nil {
		m.FailIfCalled.Fail()
		return nil, nil
	} else if m.ReturnNil {
		return nil, nil
	} else if m.ReturnErr {
		return nil, errors.Errorf("mock error")
	} else if m.ReturnBytes != nil {
		return m.ReturnBytes, nil
	} else {
		return []byte(m.ReturnString), nil
	}
}

func (m *MockTokenProvider) Reissue() ([]byte, error) {
	return m.Get()
}
