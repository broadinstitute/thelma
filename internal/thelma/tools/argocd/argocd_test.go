package argocd

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_Login(t *testing.T) {
	testCases := []struct {
		name             string
		vpnCheckResponse fakeHTTPResponse
		setupCommands    func(runner *shell.MockRunner, host string)
		expectError      string
	}{
		{
			name: "happy path",
			setupCommands: func(runner *shell.MockRunner, host string) {
				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--grpc-web",
						"--plaintext",
						"login",
						"--sso",
						host,
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", host),
					},
				})

				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--grpc-web",
						"--plaintext",
						"account",
						"get-user-info",
						"--output",
						"yaml",
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", host),
					},
				}).WithStdout("loggedIn: true\n")
			},
		},
		{
			name: "not on VPN",
			vpnCheckResponse: fakeHTTPResponse{
				status: notOnVpnResponseStatus,
				body:   "403 Forbitten",
			},
			expectError: "you must log in to the non-split VPN",
		},
		{
			name: "login command fails",
			setupCommands: func(runner *shell.MockRunner, host string) {
				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--grpc-web",
						"--plaintext",
						"login",
						"--sso",
						host,
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", host),
					},
				}).Exits(2)
			},
			expectError: "login.*exited with status 2",
		},
		{
			name: "login check fails",
			setupCommands: func(runner *shell.MockRunner, host string) {
				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--grpc-web",
						"--plaintext",
						"login",
						"--sso",
						host,
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", host),
					},
				})

				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--grpc-web",
						"--plaintext",
						"account",
						"get-user-info",
						"--output",
						"yaml",
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", host),
					},
				}).WithStdout("loggedIn: false\n")
			},
			expectError: "login command succeeded but client is not logged in",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			httpServer := NewFakeHTTPServer(t, tc.vpnCheckResponse)
			u, err := url.Parse(httpServer.URL)
			require.NoError(t, err)
			host := u.Host

			runner := shell.DefaultMockRunner()
			app, err := builder.NewBuilder().
				WithTestDefaults(t).
				SetShellRunner(runner).
				SetConfigOverride("argocd.tls", false).
				SetConfigOverride("argocd.host", host).
				Build()
			require.NoError(t, err)

			if tc.setupCommands != nil {
				tc.setupCommands(runner, host)
			}

			err = Login(app.Config(), app.ShellRunner())

			if tc.expectError == "" {
				require.NoError(t, err)
				return
			}

			assert.Error(t, err)
			assert.Regexp(t, tc.expectError, err.Error())
		})
	}
}

type fakeHTTPResponse struct {
	status int
	body   string
}

func NewFakeHTTPServer(t *testing.T, resp fakeHTTPResponse) *httptest.Server {
	if resp.status == 0 {
		resp.status = 200
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(resp.status)
		_, err := w.Write([]byte(resp.body))
		require.NoError(t, err)
	}))
	t.Cleanup(s.Close)
	return s
}
