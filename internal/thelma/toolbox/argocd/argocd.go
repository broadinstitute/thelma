package argocd

import (
	"bytes"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/broadinstitute/thelma/internal/thelma/app/credentials"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	naming "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const prog = `argocd`
const configPrefix = `argocd`
const yamlFormat = "yaml"
const applicationNamespace = "argocd"

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

// retryBaseInterval base interval for how long to wait between retries of failed ArgoCD cli commands
// (passed to backoff algo)
const retryBaseInterval = 10 * time.Second

// retryMaxDelay upper limit for retry interval for failed ArgoCD cli commands
const retryMaxDelay = time.Minute

// retryAttempts max number of attempts to run a failed ArgoCD cli command
const retryAttempts = 4

// unretryableErrors holds list of errors that indicate a failed ArgoCD cli command should NOT be re-run
var unretryableErrors = []*regexp.Regexp{
	// don't retry commands that fail with a timeout error, whatever we were waiting for (sync to finish, app to
	// become healthy) did not finish successfully within the desired timeout.
	regexp.MustCompile(regexp.QuoteMeta("rpc error: code = Canceled")),
	// don't retry commands that fail with a 401, because we actually use 401s from ArgoCD to indicate that we
	// need to re-run ArgoCD auth process -- if we retry, it seems to the user like Thelma is hanging.
	regexp.MustCompile(regexp.QuoteMeta("failed with status code 401")),
}

// SyncOptions options for an ArgoCD sync operation
type SyncOptions struct {
	// HardRefresh if true, perform a hard refresh before syncing Argo apps
	HardRefresh bool
	// SyncIfNoDiff if true, sync even if the hard refresh indicates there are no config differences
	SyncIfNoDiff bool
	// NeverSync if true, never actually sync apps to allow refresh-only behavior
	NeverSync bool
	// WaitHealthy if true, wait for the application to become healthy after syncing
	WaitHealthy bool
	// WaitHealthyTimeout how long to wait for the application to become healthy before giving up
	WaitHealthyTimeoutSeconds int
	// OnlyLabels if not empty, only sync resources with the given labels
	OnlyLabels map[string]string
	// SkipLegacyConfigsRestart if true, do not restart deployments to pick up firecloud-develop changes
	SkipLegacyConfigsRestart bool
	// StatusReporter pool.StatusReporter
	StatusReporter pool.StatusReporter
}

func (s SyncOptions) reportStatus(message string) {
	if s.StatusReporter != nil {
		s.StatusReporter.Update(pool.Status{Message: message})
	}
}

type SyncOption func(options *SyncOptions)

type WaitExistOptions struct {
	// WaitExistTimeoutSeconds how long to wait for an application to exist before timing out
	WaitExistTimeoutSeconds int `default:"300"`

	// WaitExistPollIntervalSeconds how long to wait between polling attempts while waiting for an app to exist
	WaitExistPollIntervalSeconds int `default:"5"`
}

type WaitExistOption func(options *WaitExistOptions)

type argocdConfig struct {
	// Host hostname of the ArgoCD server
	Host string `valid:"hostname" default:"argocd.dsp-devops-prod.broadinstitute.org"`

	SherlockOidcProvider string `valid:"url" default:"https://sherlock-oidc.dsp-devops-prod.broadinstitute.org/oidc"`
	// SherlockOidcCliClient is the public CLI client ID for running PKCE native OAuth2 flows from Thelma against
	// Sherlock's OIDC provider. This value is intentionally public in a manner not dissimilar from the IAP
	// client info stored within Thelma's codebase.
	SherlockOidcCliClientID string `default:"PjVR6GrFsnKMN7k9Ldo9vWrvg5zPtEMOxSgbSaewoAo"`

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
	SyncTimeoutSeconds int `default:"900"`

	// SyncRetries how many times to retry failed sync operations before giving up
	SyncRetries int `default:"4"`

	// WaitHealthyTimeoutSeconds how long to wait for an application to become healthy after syncing
	WaitHealthyTimeoutSeconds int `default:"900"`

	WaitExistOptions
}

// SyncResult stores information about the outcome of a Sync operation
type SyncResult struct {
	// Synced true if the app was actually synced, false if not
	Synced bool
}

// ArgoCD is for running `argocd` commands.
// Note: we explored using the ArgoCD golang client, but the ArgoCD API is gRPC and designed for async UI communication.
// As a result it is extremely complicated to do things that are trivial via the CLI.
type ArgoCD interface {
	// SyncApp will sync an ArgoCD app
	SyncApp(appName string, options ...SyncOption) (SyncResult, error)
	// HardRefresh will hard refresh an ArgoCD app (force a manifest re-render without a corresponding git change)
	HardRefresh(appName string) error
	// WaitExist will wait for an ArgoCD app to exist
	WaitExist(appName string, options ...WaitExistOption) error
	// SyncRelease will sync a Terra release's ArgoCD app(s), including the legacy configs app if there is one
	SyncRelease(release terra.Release, options ...SyncOption) error
	// AppStatus returns a summary of an application's health status
	AppStatus(appName string) (ApplicationStatus, error)
	// DestinationURL returns a URL to an environment's Argo applications
	DestinationURL(dest terra.Destination) string
	// DefaultSyncOptions returns default sync options
	DefaultSyncOptions() SyncOptions
}

func New(thelmaConfig config.Config, shellRunner shell.Runner, creds credentials.Credentials, iapTokenProvider credentials.TokenProvider, sherlockHttpClient *http.Client) (ArgoCD, error) {
	var err error
	var cfg argocdConfig
	if err = thelmaConfig.Unmarshal(configPrefix, &cfg); err != nil {
		return nil, err
	}
	var tokenProvider credentials.TokenProvider

	if os.Getenv(envVars.token) != "" {
		// Backwards compatibility, use env token if provided
		log.Debug().Msgf("Env var %s is set; will be used to to authenticate ArgoCD CLI commands", envVars.token)
	} else {
		tokenProvider, err = implicitTokenProvider(creds, cfg, sherlockHttpClient)
		if err != nil {
			return nil, err
		}
	}

	a := &argocd{
		runner: shellRunner,
		cfg:    cfg,

		iapTokenProvider: iapTokenProvider,
		tokenProvider:    tokenProvider,
	}

	if err = a.ensureLoggedIn(); err != nil {
		return nil, err
	}

	return a, nil
}

// implements ArgoCD interface
type argocd struct {
	runner   shell.Runner
	cfg      argocdConfig
	iapToken string
	token    string

	iapTokenProvider credentials.TokenProvider
	tokenProvider    credentials.TokenProvider
}

func (a *argocd) SyncApp(appName string, options ...SyncOption) (SyncResult, error) {
	opts := a.asSyncOptions(options...)

	var result SyncResult

	// refresh the app, using hard refresh if needed
	opts.reportStatus(fmt.Sprintf("Refreshing %s", appName))
	hasDifferences, err := a.diffWithRetries(appName, opts)
	if err != nil {
		return result, err
	}
	if !hasDifferences {
		if opts.SyncIfNoDiff {
			log.Debug().Msgf("%s is in sync, will sync anyway", appName)
		} else if opts.NeverSync {
			log.Debug().Msgf("%s is in sync, wouldn't sync due to options", appName)
			return result, err
		} else {
			log.Debug().Msgf("%s is in sync, won't trigger a new sync", appName)
			return result, err
		}
	} else if opts.NeverSync {
		log.Debug().Msgf("%s is out of sync, won't sync due to options", appName)
		return result, err
	}

	opts.reportStatus(fmt.Sprintf("Waiting in-progress %s", appName))
	if err := a.waitForInProgressSyncToComplete(appName); err != nil {
		return result, err
	}

	// we're about to sync, so update result to indicate we made an attempt
	result.Synced = true
	opts.reportStatus(fmt.Sprintf("Syncing %s", appName))
	if err := a.sync(appName, opts); err != nil {
		return result, err
	}

	if opts.WaitHealthy {
		opts.reportStatus(fmt.Sprintf("Waiting healthy %s", appName))
		if err := a.waitHealthy(appName, opts.WaitHealthyTimeoutSeconds); err != nil {
			return result, err
		}
	}

	log.Debug().Msgf("Successfully synced %s", appName)
	return result, nil
}

func (a *argocd) SyncRelease(release terra.Release, options ...SyncOption) error {
	syncOpts := a.asSyncOptions(options...)

	hasLegacyConfigsApp, err := a.hasLegacyConfigsApp(release)
	if err != nil {
		return err
	}

	legacyConfigsApp := naming.LegacyConfigsApplicationName(release)
	primaryApp := naming.ApplicationName(release)

	// Sync the legacy configs app, if one exists
	legacyConfigsWereSynced := false
	if hasLegacyConfigsApp {
		// Sherlock has dropped support for firecloud-develop refs. There also aren't any more legacy config apps,
		// but a smaller refactoring is to remove the references to firecloud-develop and just hardcode to this to
		// dev for now.
		if err := a.setRef(legacyConfigsApp, "dev"); err != nil {
			return err
		}
		syncResult, err := a.SyncApp(legacyConfigsApp, options...)
		if err != nil {
			return err
		}
		legacyConfigsWereSynced = syncResult.Synced
	}

	// Sync primary app without waiting for it to become healthy
	optionsNoWaitHealthy := append(options, func(options *SyncOptions) {
		options.WaitHealthy = false
	})

	if err := a.setRef(primaryApp, release.TerraHelmfileRef()); err != nil {
		return err
	}
	if _, err := a.SyncApp(primaryApp, optionsNoWaitHealthy...); err != nil {
		return err
	}

	if hasLegacyConfigsApp {
		if syncOpts.SkipLegacyConfigsRestart {
			log.Debug().Msgf("Won't restart deployments to pick up firecloud-develop changes (legacy config restarts are skipped)")
		} else {
			if legacyConfigsWereSynced {
				log.Debug().Msgf("Waiting for %s to become healthy before restarting deployments", primaryApp)
				syncOpts.reportStatus(fmt.Sprintf("Waiting healthy %s", primaryApp))
				if err := a.waitHealthy(primaryApp, syncOpts.WaitHealthyTimeoutSeconds); err != nil {
					return err
				}

				log.Debug().Msgf("Restarting deployments in %s to pick up potential firecloud-develop config changes", primaryApp)
				syncOpts.reportStatus(fmt.Sprintf("Restart deployments %s", primaryApp))
				if err := a.restartDeployments(primaryApp); err != nil {
					return err
				}
			} else {
				log.Debug().Msgf("No firecloud-develop changes detected, won't restart deployments")
			}
		}
	}

	// Now wait for the primary app to become healthy
	if syncOpts.WaitHealthy {
		return a.waitHealthy(primaryApp, syncOpts.WaitHealthyTimeoutSeconds)
	}
	return nil
}

func (a *argocd) HardRefresh(appName string) error {
	_, err := a.diffWithRetries(appName, a.DefaultSyncOptions())
	return err
}

func (a *argocd) AppStatus(appName string) (ApplicationStatus, error) {
	app, err := a.getApplication(appName)
	if err != nil {
		return ApplicationStatus{}, err
	}
	return app.Status, nil
}

func (a *argocd) waitHealthy(appName string, timeoutSeconds int) error {
	log.Debug().Msgf("Waiting up to %d seconds for %s to become healthy", timeoutSeconds, appName)

	return a.runCommandWithRetries([]string{
		"app",
		"wait",
		appName,
		"--timeout",
		fmt.Sprintf("%d", timeoutSeconds),
		"--health",
	})
}

func (a *argocd) WaitExist(appName string, options ...WaitExistOption) error {
	var opts WaitExistOptions
	opts.WaitExistTimeoutSeconds = a.cfg.WaitExistTimeoutSeconds
	opts.WaitExistPollIntervalSeconds = a.cfg.WaitExistPollIntervalSeconds
	for _, opt := range options {
		opt(&opts)
	}

	logger := log.With().Str("argo-app", appName).Logger()

	timeout := time.Second * (time.Duration(opts.WaitExistTimeoutSeconds))
	pollInterval := time.Second * (time.Duration(opts.WaitExistPollIntervalSeconds))

	logger.Info().Msgf("Waiting up to %s for %s to exist", timeout, appName)

	doneCh := make(chan bool, 1)
	timeoutCh := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-timeoutCh:
				logger.Debug().Msgf("Timeout reached, exiting polling")
				return
			default:
				if err := a.runCommandOnce([]string{
					"app",
					"get",
					appName,
				}); err == nil {
					log.Debug().Msgf("%s exists", appName)
					doneCh <- true
					return
				}
				log.Debug().Msgf("%s does not exist, will check again in %s", appName, pollInterval)
				time.Sleep(pollInterval)
			}
		}
	}()

	select {
	case <-doneCh:
		return nil
	case <-time.After(timeout):
		timeoutCh <- true
		return errors.Errorf("timed out after %s waiting for Argo application %s to exist", timeout, appName)
	}
}

func (a *argocd) DestinationURL(d terra.Destination) string {
	var u url.URL
	if a.cfg.TLS {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}

	u.Host = a.cfg.Host

	u.Path = "/applications"

	labels := make(map[string]string)

	// We have seen a few transient cases where the terra.Destination parameter is nil, resulting in a SIGSEGV.
	// This is a defensive check to prevent a panic in that case.
	// This function is only used for logging, so it's not a big deal if we return an empty URL.
	if d == nil {
		log.Warn().Msgf("Destination is nil when generating argocd url, returning empty URL")
		return ""
	}

	if d.IsEnvironment() {
		labels["env"] = d.Name()
	} else {
		labels["type"] = "cluster"
		labels["cluster"] = d.Name()
	}

	selector := joinSelector(labels)

	params := url.Values{
		"labels": {selector},
	}
	u.RawQuery = params.Encode()

	return u.String()
}

func (a *argocd) DefaultSyncOptions() SyncOptions {
	return SyncOptions{
		HardRefresh:               true,
		SyncIfNoDiff:              false,
		WaitHealthy:               true,
		WaitHealthyTimeoutSeconds: a.cfg.WaitHealthyTimeoutSeconds,
	}
}

func (a *argocd) asSyncOptions(options ...SyncOption) SyncOptions {
	opts := a.DefaultSyncOptions()
	for _, option := range options {
		option(&opts)
	}
	return opts
}

func (a *argocd) restartDeployments(appName string) error {
	if err := a.runCommandWithRetries([]string{
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
	return a.runCommandWithRetries([]string{"app", "actions", "run", "--kind=Deployment", appName, "restart", "--all"})
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

	legacyConfigsName := naming.LegacyConfigsApplicationName(release)
	for _, line := range lines {
		if strings.TrimSpace(line) == fmt.Sprintf("%s/%s", applicationNamespace, legacyConfigsName) {
			return true, nil
		}
	}
	return false, nil
}

func (a *argocd) waitForInProgressSyncToComplete(appName string) error {
	log.Debug().Msgf("Waiting up to %d seconds for in-progress sync operations on %s to complete", a.cfg.WaitInProgressOperationTimeoutSeconds, appName)

	return a.runCommandWithRetries([]string{
		"app",
		"wait",
		appName,
		"--operation",
		"--timeout",
		fmt.Sprintf("%d", a.cfg.WaitInProgressOperationTimeoutSeconds),
	})
}

func (a *argocd) sync(appName string, opts SyncOptions) error {
	log.Debug().Msgf("Syncing ArgoCD app: %s", appName)

	args := []string{
		"app",
		"sync",
		appName,
		"--retry-limit",
		fmt.Sprintf("%d", a.cfg.SyncRetries),
		"--prune",
		"--timeout",
		fmt.Sprintf("%d", a.cfg.SyncTimeoutSeconds),
	}

	if len(opts.OnlyLabels) > 0 {
		args = append(args, "--label", joinSelector(opts.OnlyLabels))
	}

	err := a.runCommandWithRetries(args)
	if err == nil {
		return nil
	}

	// Log a warning instead of returning an error if we got "No matching resources found"
	if strings.Contains(err.Error(), "No matching resources found for labels") {
		log.Warn().Err(err).Msgf("Selective sync failed: no matching resources")
		return nil
	}

	return err
}

func (a *argocd) diffWithRetries(appName string, opts SyncOptions) (hasDifferences bool, err error) {
	for i := 1; i <= a.cfg.DiffRetries; i++ {
		hasDifferences, err = a.diff(appName, opts)
		if err == nil {
			return hasDifferences, err
		}
		log.Warn().Str("app", appName).Int("count", i).Err(err).Msgf("attempt %d to diff %s returned error: %v", i, appName, err)

		if i < a.cfg.DiffRetries {
			log.Warn().Str("app", appName).Int("count", i).Msgf("Will retry in %s", a.cfg.DiffRetryInterval)
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
	// don't use default retries here because this command is already wrapped in retries
	// with specific, custom options that have been adjusted over time.
	err := a.runCommandOnce(args)
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
	if err != nil && !strings.Contains(err.Error(), "failed with status code 401") {
		// failed with status code 401 means auth token expired, we return a special error message (see below) in that case
		return err
	}
	if err != nil || !output.LoggedIn {
		return errors.Errorf("ArgoCD client is not authenticated; please retry or supply an ArgoCD token via %s", envVars.token)
	}
	return nil
}

// run `argocd app set <app-name> --revision=<ref>` to set an Argo app's git ref
func (a *argocd) setRef(appName string, ref string) error {
	log.Info().Msgf("Setting app %s to ref %s", appName, ref)
	err := a.runCommandWithRetries([]string{"app", "set", appName, "--revision", ref, "--validate=false"})
	if err != nil {
		return errors.Errorf("error setting %s to revision %q: %v", appName, ref, err)
	}
	return nil
}

// run `argocd app get <app-name>` to retrive an ArgoCD application's YAML definition
func (a *argocd) getApplication(appName string) (application, error) {
	var app application

	buf := bytes.Buffer{}
	err := a.runCommandWithRetries([]string{"app", "get", appName, "-o", "yaml"}, func(options *shell.RunOptions) {
		options.Stdout = &buf
	})
	if err != nil {
		return app, err
	}

	if err = yaml.Unmarshal(buf.Bytes(), &app); err != nil {
		return app, errors.Errorf("error unmarshalling argo app %s: %v", appName, err)
	}

	return app, nil
}

func (a *argocd) runCommandAndParseYamlOutput(args []string, out interface{}) error {
	buf := new(bytes.Buffer)

	args = append(args, flags.outputFormat, yamlFormat)

	err := a.runCommandWithRetries(args, func(options *shell.RunOptions) {
		options.Stdout = buf
	})
	if err != nil {
		return err
	}

	cmdOutput := buf.Bytes()
	if err := yaml.Unmarshal(cmdOutput, out); err != nil {
		return errors.Errorf("error unmarshalling command output: %v", err)
	}

	return nil
}

func (a *argocd) runCommandAndParseLineSeparatedOutput(args []string) ([]string, error) {
	buf := new(bytes.Buffer)

	err := a.runCommandWithRetries(args, func(options *shell.RunOptions) {
		options.Stdout = buf
	})
	if err != nil {
		return []string{}, err
	}

	cmdOutput := buf.String()
	return strings.Split(cmdOutput, "\n"), nil
}

func isRetryableError(err error) bool {
	exitErr, ok := err.(*shell.ExitError)
	if !ok {
		return false
	}

	for _, regex := range unretryableErrors {
		if regex.MatchString(exitErr.Stderr) {
			return false
		}
	}
	return true
}

func (a *argocd) runCommandWithRetries(args []string, options ...shell.RunOption) error {
	return retry.Do(
		func() error {
			return a.runCommandOnce(args, options...)
		},
		retry.RetryIf(isRetryableError),
		retry.Delay(retryBaseInterval),
		retry.DelayType(retry.BackOffDelay),
		retry.MaxDelay(retryMaxDelay),
		retry.Attempts(retryAttempts),
		retry.OnRetry(func(n uint, err error) {
			log.Debug().Err(err).Msgf("argocd cli command failed, will retry up to %d times", retryAttempts)
		}),
	)
}

func (a *argocd) runCommandOnce(args []string, options ...shell.RunOption) error {
	// build env var list
	var env []string
	if a.cfg.Host != "" {
		env = append(env, fmt.Sprintf("%s=%s", envVars.server, a.cfg.Host))
	}
	if token, tokenOk, err := a.authToken(); err != nil {
		return err
	} else if tokenOk {
		env = append(env, fmt.Sprintf("%s=%s", envVars.token, token))
	}

	// build arg list
	var _args []string

	// add IAP token header
	proxyAuthorizationHeader, err := a.proxyAuthorizationHeader()
	if err != nil {
		return err
	}
	_args = append(_args, flags.header, proxyAuthorizationHeader)

	if a.cfg.GRPCWeb {
		_args = append(_args, flags.grpcWeb)
	}
	if !a.cfg.TLS {
		_args = append(_args, flags.plainText)
	}

	_args = append(_args, args...)

	options = append(options, func(runOpts *shell.RunOptions) {
		// argo cli commands are extremely noisy, collect at trace level
		runOpts.OutputLogLevel = zerolog.TraceLevel
	})

	return a.runner.Run(
		shell.Command{
			Prog: prog,
			Args: _args,
			Env:  env,
		},
		options...,
	)
}

func (a *argocd) authToken() (string, bool, error) {
	if a.token != "" {
		return a.token, true, nil
	} else if a.tokenProvider != nil {
		token, err := a.tokenProvider.Get()
		if err != nil {
			return "", false, err
		}
		return string(token), true, nil
	} else {
		return "", false, nil
	}
}

func (a *argocd) proxyAuthorizationHeader() (string, error) {
	if a.iapToken != "" {
		return fmt.Sprintf("Proxy-Authorization: Bearer %s", a.iapToken), nil
	} else if a.iapTokenProvider != nil {
		token, err := a.iapTokenProvider.Get()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Proxy-Authorization: Bearer %s", string(token)), nil
	} else {
		return "", errors.New("argocd: no IAP token or token provider available")
	}
}

// releaseSelector returns set of selectors for all argo apps associated with a release
// (often just the primary application, but can include the legacy configs application as well)
func releaseSelector(release terra.Release) map[string]string {
	if release.IsAppRelease() {
		return map[string]string{
			"app": release.Name(),
			"env": release.Destination().Name(),
		}
	} else {
		return map[string]string{
			"release": release.Name(),
			"cluster": release.Destination().Name(),
			"type":    "cluster",
		}
	}
}

// joinSelector join map of label key-value pairs {"a":"b", "c":"d"} into selector string "a=b,c=d"
func joinSelector(labels map[string]string) string {
	var list []string
	for name, value := range labels {
		list = append(list, fmt.Sprintf("%s=%s", name, value))
	}
	sort.Strings(list)
	return strings.Join(list, ",")
}
