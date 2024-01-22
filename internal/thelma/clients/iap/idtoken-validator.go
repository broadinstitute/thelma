package iap

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/api/idtoken"
)

func idtokenValidator(value []byte) error {
	payload, err := idtoken.Validate(context.Background(), string(value), "")
	if err != nil {
		return errors.Errorf("failed to validate ID token JWT: %v", err)
	} else if payload == nil {
		return errors.Errorf("ID token JWT seemed to pass validation but payload was nil")
	} else {
		return nil
	}
}
