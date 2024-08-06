package serve_redirect

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/pkg/errors"
)

func generateState() (string, error) {
	stateBytes := make([]byte, 32)
	_, err := rand.Read(stateBytes)
	if err != nil {
		return "", errors.Errorf("failed to generate state: %v", err)
	}
	return base64.URLEncoding.EncodeToString(stateBytes), nil
}
