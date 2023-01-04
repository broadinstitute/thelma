package testutils

import "golang.org/x/oauth2"

type fakeTokenSource struct {
	fakeToken string
}

func NewFakeTokenSource(token string) oauth2.TokenSource {
	return &fakeTokenSource{
		fakeToken: token,
	}
}

func (f *fakeTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: f.fakeToken}, nil
}
