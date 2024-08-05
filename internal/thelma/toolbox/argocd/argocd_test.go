package argocd

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const fakeArgocdHost = "fake-argo.com"
const fakeIapToken = "fake-iap-token"
const fakeToken = "fake-token"

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
		WithStdout("argocd/leonardo-configs-dev\nargocd/leonardo-dev\n")

	// sync legacy configs app
	_mocks.expectCmd("app", "set", "leonardo-configs-dev", "--revision", "dev", "--validate=false")

	_mocks.expectCmd("app", "diff", "leonardo-configs-dev", "--hard-refresh").Exits(1) // non-zero indicates a diff was detected

	_mocks.expectCmd("app", "wait", "leonardo-configs-dev", "--operation", "--timeout", "300")

	_mocks.expectCmd("app", "sync", "leonardo-configs-dev", "--retry-limit", "4", "--prune", "--timeout", "900")

	_mocks.expectCmd("app", "wait", "leonardo-configs-dev", "--timeout", "900", "--health")

	// sync primary app
	_mocks.expectCmd("app", "set", "leonardo-dev", "--revision", "HEAD", "--validate=false")

	_mocks.expectCmd("app", "diff", "leonardo-dev", "--hard-refresh").Exits(1) // non-zero indicates a diff was detected

	_mocks.expectCmd("app", "wait", "leonardo-dev", "--operation", "--timeout", "300")

	_mocks.expectCmd("app", "sync", "leonardo-dev", "--retry-limit", "4", "--prune", "--timeout", "900")

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
		WithStdout("argocd/leonardo-configs-dev\nargocd/leonardo-dev\n")

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

func Test_ensureLoggedIn(t *testing.T) {
	_mocks := setupMocks(t)
	_argocd := _mocks.argocd

	_mocks.expectCmd("account", "get-user-info", "--output", "yaml").WithStdout("loggedIn: true")
	require.NoError(t, _argocd.ensureLoggedIn())
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
			msg: "is retryable",
			exp: true,
		},
		{
			msg: "rpc error: code = Canceled desc = context canceled",
			exp: false,
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
	fakeToken    string
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
		Env:  []string{envVars.server + "=" + m.fakeHost, envVars.token + "=" + m.fakeToken},
	})
}

func setupMocks(t *testing.T) *mocks {
	iapToken := fakeIapToken
	token := fakeToken
	host := fakeArgocdHost

	testConfig, err := config.NewTestConfig(t, map[string]interface{}{
		"argocd.host": host,
	})
	require.NoError(t, err)

	mockRunner := shell.DefaultMockRunner()

	var cfg argocdConfig
	err = testConfig.Unmarshal(configPrefix, &cfg)
	require.NoError(t, err)

	_argocd := &argocd{
		runner:   mockRunner,
		cfg:      cfg,
		iapToken: iapToken,
		token:    token,
	}

	return &mocks{
		fakeIapToken: iapToken,
		fakeToken:    token,
		fakeHost:     host,
		argocd:       _argocd,
		runner:       mockRunner,
	}
}
