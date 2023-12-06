package sherlock

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const sherlockGhaOidcHeader = "X-GHA-OIDC-JWT"

func makeClientAuthWriter(
	iapTokenProvider credentials.TokenProvider,
	ghaOidcTokenProvider credentials.TokenProvider,
) runtime.ClientAuthInfoWriter {
	mechanisms := make([]runtime.ClientAuthInfoWriter, 0, 2)

	if iapTokenProvider != nil {
		mechanisms = append(mechanisms, &iapInfoWriter{tokenProvider: iapTokenProvider})
	}

	if ghaOidcTokenProvider != nil {
		mechanisms = append(mechanisms, &ghaOidcInfoWriter{tokenProvider: ghaOidcTokenProvider})
	}

	return httptransport.Compose(mechanisms...)
}

type iapInfoWriter struct {
	tokenProvider credentials.TokenProvider
}

func (i *iapInfoWriter) AuthenticateRequest(request runtime.ClientRequest, _ strfmt.Registry) error {
	if token, err := i.tokenProvider.Get(); err != nil {
		return err
	} else if len(token) == 0 {
		return nil
	} else {
		return request.SetHeaderParam("Authorization", fmt.Sprintf("Bearer %s", string(token)))
	}
}

type ghaOidcInfoWriter struct {
	tokenProvider credentials.TokenProvider
}

func (g *ghaOidcInfoWriter) AuthenticateRequest(request runtime.ClientRequest, _ strfmt.Registry) error {
	if token, err := g.tokenProvider.Get(); err != nil {
		return err
	} else if len(token) == 0 {
		return nil
	} else {
		return request.SetHeaderParam(sherlockGhaOidcHeader, string(token))
	}
}
