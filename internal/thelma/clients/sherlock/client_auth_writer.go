package sherlock

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const sherlockGhaOidcHeader = "X-GHA-OIDC-JWT"

func makeClientAuthWriter(
	iapToken string,
	ghaOidcTokenProvider credentials.TokenProvider,
) runtime.ClientAuthInfoWriter {
	mechanisms := []runtime.ClientAuthInfoWriter{
		httptransport.BearerToken(iapToken),
	}

	if ghaOidcTokenProvider != nil {
		mechanisms = append(mechanisms, &ghaOidcInfoWriter{tokenProvider: ghaOidcTokenProvider})
	}

	return httptransport.Compose(mechanisms...)
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
