package iap

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
)

// this tokenProvider decorates the credentials.TokenProvider to extract and return the JUST identity token field to
// the client, instead of the entire OAuth token (which is what is persisted to disk).
type tokenProvider struct {
	credentials.TokenProvider
}

func (p *tokenProvider) Get() ([]byte, error) {
	serializedPersistentToken, err := p.TokenProvider.Get()
	if err != nil {
		return nil, err
	}

	return extractIdentityToken(serializedPersistentToken)
}

func (p *tokenProvider) Reissue() ([]byte, error) {
	serializedPersistentToken, err := p.TokenProvider.Reissue()
	if err != nil {
		return nil, err
	}

	return extractIdentityToken(serializedPersistentToken)
}

func extractIdentityToken(serializedPersistentToken []byte) ([]byte, error) {
	oauthToken, err := unmarshalPersistentToken(serializedPersistentToken)
	if err != nil {
		return nil, err
	}

	asString, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("error extracting id_token field (type assertion failed)")
	}

	return []byte(asString), nil
}
