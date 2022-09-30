package argocd

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const fakeArgocdHost = "fake-argo.com"

func Test_Login(t *testing.T) {
	testCases := []struct {
		name          string
		setupCommands func(runner *shell.MockRunner)
		expectError   string
	}{
		{
			name: "happy path",
			setupCommands: func(runner *shell.MockRunner) {
				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--header",
						"Proxy-Authorization: Bearer my-iap-token",
						"--grpc-web",
						"login",
						"--sso",
						fakeArgocdHost,
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", fakeArgocdHost),
					},
				})

				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--header",
						"Proxy-Authorization: Bearer my-iap-token",
						"--grpc-web",
						"account",
						"get-user-info",
						"--output",
						"yaml",
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", fakeArgocdHost),
					},
				}).WithStdout("loggedIn: true\n")
			},
		},
		{
			name: "login command fails",
			setupCommands: func(runner *shell.MockRunner) {
				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--header",
						"Proxy-Authorization: Bearer my-iap-token",
						"--grpc-web",
						"login",
						"--sso",
						fakeArgocdHost,
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", fakeArgocdHost),
					},
				}).Exits(2)
			},
			expectError: "login.*exited with status 2",
		},
		{
			name: "browserLogin check fails",
			setupCommands: func(runner *shell.MockRunner) {
				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--header",
						"Proxy-Authorization: Bearer my-iap-token",
						"--grpc-web",
						"login",
						"--sso",
						fakeArgocdHost,
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", fakeArgocdHost),
					},
				})

				runner.ExpectCmd(shell.Command{
					Prog: "argocd",
					Args: []string{
						"--header",
						"Proxy-Authorization: Bearer my-iap-token",
						"--grpc-web",
						"account",
						"get-user-info",
						"--output",
						"yaml",
					},
					Env: []string{
						fmt.Sprintf("ARGOCD_SERVER=%s", fakeArgocdHost),
					},
				}).WithStdout("loggedIn: false\n")
			},
			expectError: "login command succeeded but client is not logged in",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := shell.DefaultMockRunner()
			thelmaConfig, err := config.Load(config.WithTestDefaults(t), config.WithOverrides(map[string]interface{}{
				"argocd.host": fakeArgocdHost,
			}))
			require.NoError(t, err)

			if tc.setupCommands != nil {
				tc.setupCommands(runner)
			}

			err = BrowserLogin(thelmaConfig, runner, "my-iap-token")

			if tc.expectError == "" {
				require.NoError(t, err)
				return
			}

			assert.Error(t, err)
			assert.Regexp(t, tc.expectError, err.Error())
		})
	}
}
