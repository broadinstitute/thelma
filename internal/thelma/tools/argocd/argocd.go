package argocd

import (
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

const prog = `argocd`
const configPrefix = `argocd`

const yamlFormat = "yaml"

// envVars holds names of environment variables we pass to the `argocd` cli
var envVars = struct {
	token  string
	server string
}{
	token:  "ARGOCD_AUTH_TOKEN",
	server: "ARGOCD_SERVER",
}

// flags holds names of global CLI flags we pass to the `argocd` cli
var flags = struct {
	grpcWeb      string
	plainText    string
	header       string
	outputFormat string
}{
	grpcWeb:      "--grpc-web",
	plainText:    "--plaintext",
	header:       "--header",
	outputFormat: "--output",
}

type SyncOptions struct {
	HardRefresh  bool
	SyncIfNoDiff bool
	WaitHealthy  bool
}

type SyncOption func(options *SyncOptions)

type argocdConfig struct {
	// Host hostname of the ArgoCD server
	Host string `valid:"hostname" default:"ap-argocd.dsp-devops.broadinstitute.org"`

	// Token (optional) token to use authenticate to ArgoCD. If not supplied, alternative authentication will be used
	Token string

	// Vault (optional) pull ArgoCD token from Vault and use that to authenticate to ArgoCD. (should only be used in CI pipelines)
	Vault struct {
		Enabled bool   `default:"false"`
		Path    string `default:"secret/devops/thelma/argocd"`
		Key     string `default:"token"`
	}

	// GRPCWeb set to true to pass --grpc-web flag to all ArgoCD commands
	GRPCWeb bool `default:"true"`

	// TLS set to true to connect to ArgoCD server over TLS
	TLS bool `default:"true"`

	// DiffRetries number of times to retry failed diff
	DiffRetries int `default:"3"`

	// DiffRetryInterval how long to sleep between diff retries
	DiffRetryInterval time.Duration `default:"5s"`

	// WaitInProgressOperationTimeoutSeconds how long to wait for an in-progress sync operation to complete before attempting our own sync
	WaitInProgressOperationTimeoutSeconds int `default:"300"`

	// SyncTimeoutSeconds timeout for sync operations
	SyncTimeoutSeconds int `default:"600"`

	// SyncRetries how many times to retry failed sync operations before giving up
	SyncRetries int `default:"4"`

	// WaitHealthyTimeoutSeconds how long to wait for an application to become healthy after syncing
	WaitHealthyTimeoutSeconds int `default:"600"`
}

// ArgoCD is for running `argocd` commands.
// Note: we explored using the ArgoCD golang client, but the ArgoCD API is gRPC and designed for async UI communication.
// As a result it is extremely complicated to do things that are trivial via the CLI.
type ArgoCD interface {
	// SyncApp will sync an ArgoCD app
	SyncApp(appName string, options ...SyncOption) error
	// SyncRelease will sync a Terra release's ArgoCD app(s), including the legacy configs app if there is one
	SyncRelease(release terra.Release, options ...SyncOption) error
	// SyncReleases will sync the ArgoCD apps for multiple Terra releases in paralle
	SyncReleases(releases []terra.Release, maxParallel int, options ...SyncOption) error
}

// Login is a thin wrapper around the `argocd login --sso` command, which:
// * presents users with a web UI to log in with their GitHub SSO credentials (same flow as logging in to the ArgoCD webapp)
// * uses those SSO credentials to generate a new ArgoCD authentication token for the user's identity
// * stores the generated token in ~/.argocd/config
func Login(thelmaConfig config.Config, shellRunner shell.Runner, iapToken string) error {
	a, err := newWithoutLoginCheck(thelmaConfig, shellRunner, iapToken, nil)
	if err != nil {
		return err
	}
	if err := a.login(); err != nil {
		return err
	}
	if err := a.ensureLoggedIn(); err != nil {
		return fmt.Errorf("error generating login creds for ArgoCD: login command succeeded but client is not logged in")
	}
	return nil
}

// New return a new ArgoCD client
func New(thelmaConfig config.Config, shellRunner shell.Runner, iapToken string, vaultClient *vaultapi.Client) (ArgoCD, error) {
	a, err := newWithoutLoginCheck(thelmaConfig, shellRunner, iapToken, vaultClient)
	if err != nil {
		return nil, err
	}
	if err := a.ensureLoggedIn(); err != nil {
		return nil, err
	}
	return a, nil
}

func newWithoutLoginCheck(thelmaConfig config.Config, shellRunner shell.Runner, iapToken string, vaultClient *vaultapi.Client) (*argocd, error) {
	var cfg argocdConfig
	if err := thelmaConfig.Unmarshal(configPrefix, &cfg); err != nil {
		return nil, err
	}

	// load token from vault if enabled and token not already configured
	if cfg.Token == "" && vaultClient != nil && cfg.Vault.Enabled {
		token, err := readTokenFromVault(cfg, vaultClient)
		if err != nil {
			return nil, err
		}
		cfg.Token = token
	}

	if cfg.Token != "" {
		logging.MaskSecret(cfg.Token)
	}

	a := &argocd{
		runner:   shellRunner,
		cfg:      cfg,
		iapToken: iapToken,
	}

	return a, nil
}

// implements ArgoCD interface
type argocd struct {
	runner   shell.Runner
	cfg      argocdConfig
	iapToken string
}

func (a *argocd) defaultSyncOptions() SyncOptions {
	return SyncOptions{
		HardRefresh:  true,
		SyncIfNoDiff: true,
		WaitHealthy:  true,
	}
}

func (a *argocd) SyncApp(appName string, options ...SyncOption) error {
	opts := a.defaultSyncOptions()
	for _, option := range options {
		option(&opts)
	}

	// refresh the app, using hard refresh if needed
	hasDifferences, err := a.diffWithRetries(appName, opts)
	if err != nil {
		return err
	}
	if !hasDifferences {
		if opts.SyncIfNoDiff {
			log.Debug().Msgf("%s is in sync, will sync anyway for side-effects", appName)
		} else {
			log.Debug().Msgf("%s is in sync, won't trigger a new sync", appName)
			return nil
		}
	}

	if err := a.waitForInProgressSyncToComplete(appName); err != nil {
		return err
	}
	if err := a.sync(appName); err != nil {
		return err
	}

	if opts.WaitHealthy {
		if err := a.waitHealthy(appName); err != nil {
			return err
		}
	}

	log.Debug().Msgf("Successfully synced %s", appName)
	return nil
}

func (a *argocd) SyncRelease(release terra.Release, options ...SyncOption) error {
	hasLegacyConfigsApp, err := a.hasLegacyConfigsApp(release)
	if err != nil {
		return err
	}

	legacyConfigsApp := LegacyConfigsApplicationName(release)
	primaryApp := ApplicationName(release)

	if hasLegacyConfigsApp {
		if err := a.SyncApp(legacyConfigsApp, options...); err != nil {
			return err
		}
	}

	if err := a.SyncApp(primaryApp, options...); err != nil {
		return err
	}

	if hasLegacyConfigsApp {
		log.Info().Msgf("Restarting deployments in %s to pick up potential firecloud-develop config changes", primaryApp)
		return a.restartDeployments(primaryApp)
	}

	return nil
}

func (a *argocd) SyncReleases(releases []terra.Release, maxParallel int, options ...SyncOption) error {
	var jobs []pool.Job
	for _, release := range releases {
		r := release
		jobs = append(jobs, pool.Job{
			Description: ApplicationName(r),
			Run: func() error {
				log.Info().Msgf("Syncing ArgoCD application(s) for %s in %s", r.Name(), r.Destination().Name())
				return a.SyncRelease(r, options...)
			},
		})
	}

	_pool := pool.New(jobs, func(options *pool.Options) {
		options.NumWorkers = maxParallel
		options.StopProcessingOnError = false
	})
	return _pool.Execute()
}

func (a *argocd) restartDeployments(appName string) error {
	if err := a.runCommand([]string{
		"app",
		"actions",
		"list",
		"--kind=Deployment",
		appName,
	}); err != nil {
		exitErr, ok := err.(*shell.ExitError)
		if !ok {
			return err
		}
		if strings.Contains(exitErr.Stderr, "No matching resource found") {
			log.Debug().Msgf("No deployments found in %s, won't attempt a restart", appName)
			return nil
		} else {
			return err
		}
	}

	log.Debug().Msgf("Restarting all deployments in %s", appName)
	return a.runCommand([]string{"app", "actions", "run", "--kind=Deployment", appName, "restart", "--all"})
}

func (a *argocd) hasLegacyConfigsApp(release terra.Release) (bool, error) {
	lines, err := a.runCommandAndParseLineSeparatedOutput([]string{
		"app",
		"list",
		"--output",
		"name",
		"--selector",
		joinSelector(releaseSelector(release)),
	})
	if err != nil {
		return false, err
	}

	legacyConfigsName := LegacyConfigsApplicationName(release)
	for _, line := range lines {
		if strings.TrimSpace(line) == legacyConfigsName {
			return true, nil
		}
	}
	return false, nil
}

func (a *argocd) waitForInProgressSyncToComplete(appName string) error {
	log.Debug().Msgf("Waiting up to %d seconds for in-progress sync operations on %s to complete", a.cfg.WaitInProgressOperationTimeoutSeconds, appName)

	return a.runCommand([]string{
		"app",
		"wait",
		appName,
		"--operation",
		"--timeout",
		fmt.Sprintf("%d", a.cfg.WaitInProgressOperationTimeoutSeconds),
	})
}

func (a *argocd) sync(appName string) error {
	log.Debug().Msgf("Syncing ArgoCD app: %s", appName)

	return a.runCommand([]string{
		"app",
		"sync",
		appName,
		"--retry-limit",
		fmt.Sprintf("%d", a.cfg.SyncRetries),
		"--prune",
		"--timeout",
		fmt.Sprintf("%d", a.cfg.SyncTimeoutSeconds),
	})
}

func (a *argocd) waitHealthy(appName string) error {
	log.Debug().Msgf("Waiting up to %d seconds for %s to become healthy", a.cfg.WaitHealthyTimeoutSeconds, appName)

	return a.runCommand([]string{
		"app",
		"wait",
		appName,
		"--timeout",
		fmt.Sprintf("%d", a.cfg.WaitHealthyTimeoutSeconds),
	})
}

func (a *argocd) diffWithRetries(appName string, opts SyncOptions) (hasDifferences bool, err error) {
	for i := 1; i <= a.cfg.DiffRetries; i++ {
		hasDifferences, err = a.diff(appName, opts)
		if err == nil {
			return hasDifferences, err
		}
		log.Warn().Str("app", appName).Int("count", i).Err(err).Msgf("attempt %d to diff %s returned error: %v", i, appName, err)

		if i < a.cfg.DiffRetries {
			log.Warn().Str("app", appName).Int("count", i).Msgf("Will retry in %d seconds", a.cfg.DiffRetryInterval)
			time.Sleep(a.cfg.DiffRetryInterval)
		}
	}
	return hasDifferences, err
}

// Run argocd diff with --hard-refresh
// Returns true if differences are found, false otherwise
func (a *argocd) diff(appName string, opts SyncOptions) (bool, error) {
	args := []string{
		"app",
		"diff",
		appName,
	}
	if opts.HardRefresh {
		args = append(args, "--hard-refresh")
	}
	err := a.runCommand(args)
	if err == nil {
		// no error means no differences were found
		return false, nil
	}

	exitErr, ok := err.(*shell.ExitError)
	if !ok {
		// weird/unexpected error, we never even ran the command
		return false, err
	}
	if exitErr.ExitCode == 1 {
		// differences were found, cool
		return true, nil
	}
	// command exited with code other than 0 or 1, so an unexpected err occurred
	return false, err
}

// run `argocd account get-user-info` and verify the output contains `loggedIn: true`
func (a *argocd) ensureLoggedIn() error {
	var output struct {
		LoggedIn bool `yaml:"loggedIn"`
	}
	err := a.runCommandAndParseYamlOutput([]string{"account", "get-user-info"}, &output)
	if err != nil {
		return err
	}
	if !output.LoggedIn {
		return fmt.Errorf("ArgoCD client is not authenticated; please run `thelma auth argocd` or supply an ArgoCD token via THELMA_ARGOCD_TOKEN")
	}
	return nil
}

// run `argocd login`
func (a *argocd) login() error {
	return a.runCommand([]string{"login", "--sso", a.cfg.Host})
}

func (a *argocd) runCommandAndParseYamlOutput(args []string, out interface{}) error {
	buf := new(bytes.Buffer)

	args = append(args, flags.outputFormat, yamlFormat)

	err := a.runCommand(args, func(options *shell.RunOptions) {
		options.Stdout = buf
	})
	if err != nil {
		return err
	}

	cmdOutput := buf.Bytes()
	if err := yaml.Unmarshal(cmdOutput, out); err != nil {
		return fmt.Errorf("error unmarshalling command output: %v", err)
	}

	return nil
}

func (a *argocd) runCommandAndParseLineSeparatedOutput(args []string) ([]string, error) {
	buf := new(bytes.Buffer)

	err := a.runCommand(args, func(options *shell.RunOptions) {
		options.Stdout = buf
	})
	if err != nil {
		return []string{}, err
	}

	cmdOutput := buf.String()
	return strings.Split(cmdOutput, "\n"), nil
}

func (a *argocd) runCommand(args []string, options ...shell.RunOption) error {
	// build env var list
	var env []string
	if a.cfg.Host != "" {
		env = append(env, fmt.Sprintf("%s=%s", envVars.server, a.cfg.Host))
	}
	if a.cfg.Token != "" {
		env = append(env, fmt.Sprintf("%s=%s", envVars.token, a.cfg.Token))
	}

	// build arg list
	var _args []string

	// add IAP token header
	_args = append(_args, flags.header, a.proxyAuthorizationHeader())

	if a.cfg.GRPCWeb {
		_args = append(_args, flags.grpcWeb)
	}
	if !a.cfg.TLS {
		_args = append(_args, flags.plainText)
	}

	_args = append(_args, args...)

	return a.runner.Run(
		shell.Command{
			Prog: prog,
			Args: _args,
			Env:  env,
		},
		options...,
	)
}

func (a *argocd) proxyAuthorizationHeader() string {
	return fmt.Sprintf("Proxy-Authorization: Bearer %s", a.iapToken)
}

func readTokenFromVault(cfg argocdConfig, vaultClient *vaultapi.Client) (string, error) {
	log.Debug().Msgf("Attempting to read ArgoCD token from Vault (%s)", cfg.Vault.Path)
	secret, err := vaultClient.Logical().Read(cfg.Vault.Path)
	if err != nil {
		return "", fmt.Errorf("error loading ArgoCD token from Vault path %s: %v", cfg.Vault.Path, err)
	}
	v, exists := secret.Data[cfg.Vault.Key]
	if !exists {
		return "", fmt.Errorf("error loading ArgoCD token from Vault path %s: missing key %s", cfg.Vault.Path, cfg.Vault.Key)
	}
	asStr, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("error loading ArgoCD token from Vault path %s: expected string key value for %s", cfg.Vault.Path, cfg.Vault.Key)
	}
	return asStr, nil
}
