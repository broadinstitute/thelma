package argocd

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const fakeArgocdHost = "fake-argo.com"
const fakeIapToken = "fake-iap-token"

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

func Test_releaseSelector(t *testing.T) {
	samDev := statemocks.NewRelease(t)
	samDev.EXPECT().IsAppRelease().Return(true)
	samDev.EXPECT().Name().Return("sam")

	dev := statemocks.NewEnvironment(t)
	dev.EXPECT().Name().Return("dev")
	samDev.EXPECT().Destination().Return(dev)

	yaleTerraDev := statemocks.NewRelease(t)
	yaleTerraDev.EXPECT().IsAppRelease().Return(false)
	yaleTerraDev.EXPECT().Name().Return("yale")

	terraDev := statemocks.NewCluster(t)
	terraDev.EXPECT().Name().Return("terra-dev")
	yaleTerraDev.EXPECT().Destination().Return(terraDev)

	assert.Equal(t, map[string]string{"app": "sam", "env": "dev"}, releaseSelector(samDev))
	assert.Equal(t, map[string]string{"release": "yale", "cluster": "terra-dev", "type": "cluster"}, releaseSelector(yaleTerraDev))
}

func Test_joinSelector(t *testing.T) {
	assert.Equal(t, "a=b,c=d,x=y", joinSelector(map[string]string{"x": "y", "a": "b", "c": "d"}))
}

func Test_SyncRelease(t *testing.T) {
	// TODO we should add more test cases with different options and config parameters
	_mocks := setupMocks(t)
	_argocd := _mocks.argocd

	dev := statemocks.NewEnvironment(t)
	dev.EXPECT().Name().Return("dev")

	leonardoDev := statemocks.NewAppRelease(t)
	leonardoDev.EXPECT().Destination().Return(dev)
	leonardoDev.EXPECT().Name().Return("leonardo")
	leonardoDev.EXPECT().IsAppRelease().Return(true)
	leonardoDev.EXPECT().TerraHelmfileRef().Return("HEAD")

	// check for legacy configs app
	_mocks.expectCmd("app", "list", "--output", "name", "--selector", "app=leonardo,env=dev").
		WithStdout("ap-argocd/leonardo-configs-dev\nap-argocd/leonardo-dev\n")

	// sync legacy configs app
	_mocks.expectCmd("app", "set", "leonardo-configs-dev", "--revision", "dev", "--validate=false")

	_mocks.expectCmd("app", "diff", "leonardo-configs-dev", "--hard-refresh").Exits(1) // non-zero indicates a diff was detected

	_mocks.expectCmd("app", "wait", "leonardo-configs-dev", "--operation", "--timeout", "300")

	_mocks.expectCmd("app", "sync", "leonardo-configs-dev", "--retry-limit", "4", "--prune", "--timeout", "600")

	_mocks.expectCmd("app", "wait", "leonardo-configs-dev", "--timeout", "900", "--health")

	// sync primary app
	_mocks.expectCmd("app", "set", "leonardo-dev", "--revision", "HEAD", "--validate=false")

	_mocks.expectCmd("app", "diff", "leonardo-dev", "--hard-refresh").Exits(1) // non-zero indicates a diff was detected

	_mocks.expectCmd("app", "wait", "leonardo-dev", "--operation", "--timeout", "300")

	_mocks.expectCmd("app", "sync", "leonardo-dev", "--retry-limit", "4", "--prune", "--timeout", "600")

	_mocks.expectCmd("app", "wait", "leonardo-dev", "--timeout", "900", "--health")

	// restart deployments
	_mocks.expectCmd("app", "actions", "list", "--kind=Deployment", "leonardo-dev")

	_mocks.expectCmd("app", "actions", "run", "--kind=Deployment", "leonardo-dev", "restart", "--all")

	_mocks.expectCmd("app", "wait", "leonardo-dev", "--timeout", "900", "--health")

	require.NoError(t, _argocd.SyncRelease(leonardoDev))
}

func Test_RefreshRelease(t *testing.T) {
	_mocks := setupMocks(t)
	_argocd := _mocks.argocd

	dev := statemocks.NewEnvironment(t)
	dev.EXPECT().Name().Return("dev")

	leonardoDev := statemocks.NewAppRelease(t)
	leonardoDev.EXPECT().Destination().Return(dev)
	leonardoDev.EXPECT().Name().Return("leonardo")
	leonardoDev.EXPECT().IsAppRelease().Return(true)
	leonardoDev.EXPECT().TerraHelmfileRef().Return("HEAD")

	// check for legacy configs app
	_mocks.expectCmd("app", "list", "--output", "name", "--selector", "app=leonardo,env=dev").
		WithStdout("ap-argocd/leonardo-configs-dev\nap-argocd/leonardo-dev\n")

	// sync legacy configs app
	_mocks.expectCmd("app", "set", "leonardo-configs-dev", "--revision", "dev", "--validate=false")

	_mocks.expectCmd("app", "diff", "leonardo-configs-dev", "--hard-refresh").Exits(1) // non-zero indicates a diff was detected

	// sync primary app
	_mocks.expectCmd("app", "set", "leonardo-dev", "--revision", "HEAD", "--validate=false")

	_mocks.expectCmd("app", "diff", "leonardo-dev", "--hard-refresh").Exits(1) // non-zero indicates a diff was detected

	require.NoError(t, _argocd.SyncRelease(leonardoDev, func(options *SyncOptions) {
		options.NeverSync = true
		options.WaitHealthy = false
		options.SkipLegacyConfigsRestart = true
	}))
}

func Test_setRef(t *testing.T) {
	_mocks := setupMocks(t)
	_argocd := _mocks.argocd

	_mocks.expectCmd("app", "set", "fake-app", "--revision", "main", "--validate=false")
	require.NoError(t, _argocd.setRef("fake-app", "main"))
}

func Test_isRetryableError(t *testing.T) {
	testCases := []struct {
		msg string
		exp bool
	}{
		{
			msg: "not retryable",
			exp: false,
		},
		{
			msg: "error communicating with server: EOF",
			exp: true,
		},
		{
			msg: "dial tcp 140.82.113.3:443: i/o timeout (Client.Timeout exceeded while awaiting headers)",
			exp: true,
		},
		{
			msg: `rpc error: code = Unknown desc = Post "https://ap-argocd.dsp-devops.broadinstitute.org:443/application.ApplicationService/Get": "dial tcp: lookup ap-argocd.dsp-devops.broadinstitute.org on 169.254.169.254:53: read udp 172.17.0.1:59204->169.254.169.254:53: i/o timeout"`,
			exp: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			err := &shell.ExitError{Stderr: tc.msg}
			assert.Equal(t, tc.exp, isRetryableError(err))
		})
	}
}

type mocks struct {
	fakeIapToken string
	fakeHost     string
	argocd       *argocd
	runner       *shell.MockRunner
}

func (m *mocks) expectCmd(args ...string) *shell.Call {
	_args := []string{"--header", "Proxy-Authorization: Bearer " + m.fakeIapToken, "--grpc-web"}
	_args = append(_args, args...)

	return m.runner.ExpectCmd(shell.Command{
		Prog: prog,
		Args: _args,
		Env:  []string{envVars.server + "=" + m.fakeHost},
	})
}

func setupMocks(t *testing.T) *mocks {
	iapToken := fakeIapToken
	host := fakeArgocdHost

	testConfig, err := config.NewTestConfig(t, map[string]interface{}{
		"argocd.host": host,
	})
	require.NoError(t, err)

	mockRunner := shell.DefaultMockRunner()
	mockRunner.ExpectCmd(shell.Command{
		Prog: prog,
		Args: []string{"--header", "Proxy-Authorization: Bearer " + iapToken, "--grpc-web", "account", "get-user-info", "--output", "yaml"},
		Env:  []string{envVars.server + "=" + host},
	}).WithStdout("loggedIn: true")

	_argocd, err := newArgocd(testConfig, mockRunner, iapToken, nil)
	require.NoError(t, err)

	return &mocks{
		fakeIapToken: iapToken,
		fakeHost:     host,
		argocd:       _argocd,
		runner:       mockRunner,
	}
}
