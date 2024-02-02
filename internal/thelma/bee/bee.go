package bee

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/bee/cleanup"
	"github.com/broadinstitute/thelma/internal/thelma/clients/slack"
	"github.com/broadinstitute/thelma/internal/thelma/ops"
	"github.com/broadinstitute/thelma/internal/thelma/ops/artifacts"
	"github.com/broadinstitute/thelma/internal/thelma/ops/logs"
	"github.com/broadinstitute/thelma/internal/thelma/ops/status"
	"github.com/pkg/errors"
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/bee/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	argocd_names "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	"github.com/rs/zerolog/log"
)

const generatorArgoApp = "terra-bee-generator"
const beeDocLink = "https://broadworkbench.atlassian.net/wiki/spaces/IA/pages/2839576631/How+to+BEE"

type Bees interface {
	DeleteWith(name string, options DeleteOptions) (*Bee, error)
	CreateWith(options CreateOptions) (*Bee, error)
	ProvisionWith(name string, options ProvisionOptions) (*Bee, error)
	SyncWith(name string, options ProvisionExistingOptions) (*Bee, error)
	StartStopWith(name string, offline bool, options StartStopOptions) (*Bee, error)
	GetBee(name string) (terra.Environment, error)
	GetTemplate(templateName string) (terra.Environment, error)
	Seeder() seed.Seeder
	FilterBees(filter terra.EnvironmentFilter) ([]terra.Environment, error)
	PinVersions(bee terra.Environment, overrides PinOptions) (terra.Environment, error)
	UnpinVersions(bee terra.Environment) error
	SyncEnvironmentGenerator(env terra.Environment) error
	SyncArgoAppsIn(env terra.Environment, options ...argocd.SyncOption) (map[terra.Release]*status.Status, error)
	ResetStatefulSets(env terra.Environment) (map[terra.Release]*status.Status, error)
	RefreshBeeGenerator() error
}

type DeleteOptions struct {
	Unseed     bool
	ExportLogs bool
}

type CreateOptions struct {
	Template string
	terra.CreateOptions
	ProvisionOptions
}

type ProvisionOptions struct {
	PinOptions          PinOptions
	Seed                bool
	SeedOptions         seed.SeedOptions
	ExportLogsOnFailure bool
	ProvisionExistingOptions
}

type ProvisionExistingOptions struct {
	SyncGeneratorOnly        bool
	WaitHealthy              bool
	WaitHealthTimeoutSeconds int
	Notify                   bool
}

type PinOptions struct {
	// Flags holds global-to-the-environment overrides like --terra-helmfile-ref, firecloud-develop-ref, --build-number
	Flags struct {
		// TerraHelmfileRef the ref the environments Argo app generator should use
		TerraHelmfileRef string
		// FirecloudDevelopRef the ref the environments Argo app generator should use
		FirecloudDevelopRef string
	}
	// FileOverrides holds overrides for individual releases, loaded from a YAML or JSON file
	FileOverrides map[string]terra.VersionOverride
}

type StartStopOptions struct {
	Notify bool
	Sync   bool
}

// Bee encapsulates operational information about a BEE
type Bee struct {
	Environment      terra.Environment
	Status           map[terra.Release]*status.Status
	ContainerLogsURL string
}

func NewBees(argocd argocd.ArgoCD, stateLoader terra.StateLoader, seeder seed.Seeder, cleanup cleanup.Cleanup, kubectl kubectl.Kubectl, ops ops.Ops, slack slack.Slack) (Bees, error) {
	state, err := stateLoader.Load()
	if err != nil {
		return nil, err
	}

	return &bees{
		argocd:      argocd,
		state:       state,
		stateLoader: stateLoader,
		seeder:      seeder,
		cleanup:     cleanup,
		kubectl:     kubectl,
		ops:         ops,
		slack:       slack,
	}, nil
}

// implements Bees interface
type bees struct {
	argocd      argocd.ArgoCD
	state       terra.State
	stateLoader terra.StateLoader
	seeder      seed.Seeder
	kubectl     kubectl.Kubectl
	cleanup     cleanup.Cleanup
	ops         ops.Ops
	slack       slack.Slack
}

func (b *bees) CreateWith(options CreateOptions) (*Bee, error) {
	template, err := b.GetTemplate(options.Template)

	if err != nil {
		return nil, err
	}

	envName, err := b.state.Environments().CreateFromTemplate(template, options.CreateOptions)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Created new environment %s", envName)

	// Reload state; required since "creating an environment" just returns the name of what was created.
	if err = b.reloadState(); err != nil {
		return nil, err
	}
	return b.ProvisionWith(envName, options.ProvisionOptions)
}

func (b *bees) ProvisionWith(name string, options ProvisionOptions) (*Bee, error) {
	bee, err := b.provisionBee(name, options)

	if bee != nil && options.Notify {
		b.trySendBeeProvisionNotification(bee.Environment, err)
	}

	return bee, err
}

func (b *bees) SyncWith(name string, options ProvisionExistingOptions) (*Bee, error) {
	env, err := b.state.Environments().Get(name)
	if err != nil {
		return nil, err
	}
	bee := &Bee{
		Environment: env,
	}
	err = b.provisionBeeApps(bee, options)
	if options.Notify && env.Owner() != "" && b.slack != nil {
		var outcome string
		if err == nil {
			outcome = "has been synced"
		} else {
			outcome = "failed to sync"
		}
		if slackErr := b.slack.SendDirectMessage(env.Owner(), fmt.Sprintf("Your <https://broad.io/beehive/r/environment/%s|%s> BEE %s.", env.Name(), env.Name(), outcome)); slackErr != nil {
			log.Warn().Msgf("Wasn't able to notify %s: %v", env.Owner(), slackErr)
		}
	}
	return bee, err
}

func (b *bees) provisionBee(name string, options ProvisionOptions) (*Bee, error) {
	env, err := b.state.Environments().Get(name)
	if err != nil {
		return nil, err
	}
	if env == nil {
		// don't think this could ever happen, but let's provide a useful error anyway
		return nil, errors.Errorf("error provisioning environment %q: missing from state", name)
	}

	bee := &Bee{
		Environment: env,
	}

	env, err = b.PinVersions(env, options.PinOptions)
	if err != nil {
		return bee, errors.Errorf("error pinning versions for environment %q: %v", name, err)
	}
	if env == nil {
		// don't think this could ever happen, but let's provide a useful error anyway
		return bee, errors.Errorf("error pinning versions for environment %q: returned env is nil?", name)
	}
	bee.Environment = env

	if err = b.provisionBeeNamespaceAndGenerator(bee); err != nil {
		return bee, err
	}

	err = b.provisionBeeAppsAndSeed(bee, options)

	if err != nil && options.ExportLogsOnFailure {
		_, logErr := b.exportLogs(bee)
		if logErr != nil {
			log.Error().Err(logErr).Msgf("error exporting logs from %s", name)
		}
		bee.ContainerLogsURL = artifacts.DefaultArtifactsURL(bee.Environment)
	}

	return bee, err
}

func (b *bees) trySendBeeProvisionNotification(env terra.Environment, maybeErr error) {
	if env.Owner() == "" {
		log.Debug().Msgf("Wanted to notify but the environment lacked an owner")
		return
	}

	if b.slack == nil {
		log.Debug().Msgf("Would have tried to notify but Slack client wasn't present; perhaps it errored earlier")
		return
	}

	log.Info().Msgf("Notifying %s", env.Owner())

	var markdown string
	if maybeErr != nil {
		// If you try to actually include the error here, Slack will try to parse it and it'll be quite unhappy.
		markdown = fmt.Sprintf("Your <https://broad.io/beehive/r/environment/%s|%s> BEE didn't come up properly; see the link and contact <#CADM7MZ35> for more information.", env.Name(), env.Name())
	} else {
		markdown = fmt.Sprintf("Your <https://broad.io/beehive/r/environment/%s|%s> BEE is ready to go!", env.Name(), env.Name())
		for _, release := range env.Releases() {
			if release.IsAppRelease() && release.ChartName() == "terraui" {
				if terraui, ok := release.(terra.AppRelease); ok {
					markdown += fmt.Sprintf(" Terra's UI is at %s.", terraui.URL())
				}
			}
		}
		markdown += fmt.Sprintf(" You'll probably want to set up your BEE with a billing account, <%s|instructions available here>.", beeDocLink)
	}

	if err := b.slack.SendDirectMessage(env.Owner(), markdown); err != nil {
		log.Warn().Msgf("Wasn't able to notify %s: %v", env.Owner(), err)
	}
}

func (b *bees) provisionBeeNamespaceAndGenerator(bee *Bee) error {
	env := bee.Environment

	if err := b.kubectl.CreateNamespace(env); err != nil {
		return errors.Errorf("error creating namespace for environment %q: %v", env.Name(), err)
	}

	if err := b.RefreshBeeGenerator(); err != nil {
		return errors.Errorf("error refreshing bee generator: %v", err)
	}

	if err := b.argocd.WaitExist(argocd_names.GeneratorName(env)); err != nil {
		return errors.Errorf("error waiting for environment generator for %s to exist: %v", env.Name(), err)
	}

	return nil
}

func (b *bees) provisionBeeAppsAndSeed(bee *Bee, options ProvisionOptions) error {
	env := bee.Environment
	if err := b.provisionBeeApps(bee, options.ProvisionExistingOptions); err != nil {
		return err
	}

	if options.Seed {
		log.Info().Msgf("Seeding BEE with test data")
		if err := b.seeder.Seed(env, options.SeedOptions); err != nil {
			return errors.Errorf("error seeding environment %q: %v", env.Name(), err)
		}
	}

	return nil
}

func (b *bees) provisionBeeApps(bee *Bee, options ProvisionExistingOptions) error {
	if err := b.SyncEnvironmentGenerator(bee.Environment); err != nil {
		return errors.Errorf("error syncing environment generator for %s: %v", bee.Environment.Name(), err)
	}
	if options.SyncGeneratorOnly {
		log.Warn().Msgf("Won't sync Argo apps for %s", bee.Environment.Name())
		return nil
	}

	log.Info().Msgf("Syncing all Argo apps in environment %s", bee.Environment.Name())
	statuses, err := b.SyncArgoAppsIn(bee.Environment, func(_options *argocd.SyncOptions) {
		// No need to do a legacy configs restart when we're changing the structure of a BEE -- we're not
		// intending to really be syncing existing chart releases
		_options.SyncIfNoDiff = true
		_options.SkipLegacyConfigsRestart = true
		_options.WaitHealthy = options.WaitHealthy
		_options.WaitHealthyTimeoutSeconds = options.WaitHealthTimeoutSeconds
	})

	bee.Status = statuses
	return err
}

func (b *bees) exportLogs(bee *Bee) (map[terra.Release]artifacts.Location, error) {
	return b.ops.Logs().Export(bee.Environment.Releases(), func(opts *logs.ExportOptions) {
		opts.Artifacts.Upload = true
	})
}

func (b *bees) DeleteWith(name string, options DeleteOptions) (*Bee, error) {
	env, err := b.GetBee(name)
	if err != nil {
		return nil, err
	}
	if env.PreventDeletion() {
		return nil, errors.Errorf("won't delete environment %s, deletion protection is enabled", env.Name())
	}

	bee := &Bee{
		Environment: env,
	}

	if options.ExportLogs {
		_, err := b.exportLogs(bee)
		if err != nil {
			log.Warn().Msgf("Container log export failed")
		}
		bee.ContainerLogsURL = artifacts.DefaultArtifactsURL(env)
	}

	if options.Unseed {
		log.Info().Msgf("Unseeding BEE before deletion")
		if err = b.seeder.Unseed(env, seed.UnseedOptions{
			Step1UnregisterAllUsers: true,
		}); err != nil {
			log.Warn().Err(err).Msgf("Failed to unseed %s; will proceed with deletion", name)
		}
	}

	if err = b.kubectl.DeleteNamespace(env); err != nil {
		return bee, err
	}

	if err = b.cleanup.Cleanup(env); err != nil {
		return bee, err
	}

	if err = b.state.Environments().Delete(env.Name()); err != nil {
		return bee, err
	}

	log.Info().Msgf("Deleted environment %s from state", name)

	log.Info().Msgf("Deleting Argo apps for %s", name)
	if err = b.RefreshBeeGenerator(); err != nil {
		return bee, err
	}

	return bee, nil
}

func (b *bees) StartStopWith(name string, offline bool, options StartStopOptions) (*Bee, error) {
	var stateDescription string
	if offline {
		stateDescription = "stopped"
	} else {
		stateDescription = "started"
	}

	if err := b.state.Environments().SetOffline(name, offline); err != nil {
		return nil, err
	}
	if err := b.reloadState(); err != nil {
		return nil, err
	}

	env, err := b.state.Environments().Get(name)
	if err != nil {
		return nil, err
	}
	if env == nil {
		// don't think this could ever happen, but let's provide a useful error anyway
		return nil, errors.Errorf("error re-loading environment %q: missing from state", env.Name())
	}

	bee := &Bee{
		Environment: env,
	}

	if options.Sync {
		statuses, err := b.SyncArgoAppsIn(env, func(options *argocd.SyncOptions) {
			options.SkipLegacyConfigsRestart = true
		})
		bee.Status = statuses
		if err != nil {
			return bee, err
		}
	}
	if options.Notify && env.Owner() != "" && b.slack != nil {
		markdown := fmt.Sprintf("Your <https://broad.io/beehive/r/environment/%s|%s> BEE has been %s", env.Name(), env.Name(), stateDescription)
		if err := b.slack.SendDirectMessage(env.Owner(), markdown); err != nil {
			log.Warn().Msgf("Wasn't able to notify %s: %v", env.Owner(), err)
		}
	}
	log.Info().Msgf("%s (https://broad.io/beehive/r/environment/%s) is now %s", env.Name(), env.Name(), stateDescription)
	return bee, nil
}

func (b *bees) SyncEnvironmentGenerator(env terra.Environment) error {
	appName := argocd_names.GeneratorName(env)
	log.Info().Msgf("Syncing generator %s for %s", appName, env.Name())
	_, err := b.argocd.SyncApp(appName)
	return err
}

func (b *bees) SyncArgoAppsIn(env terra.Environment, options ...argocd.SyncOption) (map[terra.Release]*status.Status, error) {
	allReleases, err := b.state.Releases().All()
	if err != nil {
		return nil, err
	}
	releases := filter.Releases().BelongsToEnvironment(env).Filter(allReleases)

	_sync, err := b.ops.Sync()
	if err != nil {
		return nil, err
	}
	return _sync.Sync(releases, len(releases), options...)
}

func (b *bees) RefreshBeeGenerator() error {
	log.Info().Msgf("Refreshing %s", generatorArgoApp)
	// workaround for a bug in ArgoCD:
	//   https://github.com/argoproj/argo-cd/issues/4505#issuecomment-880271371
	// We perform a hard refresh with autosync
	return b.argocd.HardRefresh(generatorArgoApp)
}

func (b *bees) PinVersions(bee terra.Environment, pinOptions PinOptions) (terra.Environment, error) {
	// pin global terra-helmfile ref, if one is specified
	if pinOptions.Flags.TerraHelmfileRef != "" {
		was := bee.TerraHelmfileRef()
		if err := b.state.Environments().PinEnvironmentToTerraHelmfileRef(bee.Name(), pinOptions.Flags.TerraHelmfileRef); err != nil {
			return nil, err
		}
		log.Info().Msgf("Set terra-helmfile ref to %s for %s (was: %s)", pinOptions.Flags.TerraHelmfileRef, bee.Name(), was)
	}

	// now, pin version overrides for individual releases.
	releaseOverrides := make(map[string]terra.VersionOverride)
	for _, r := range bee.Releases() {
		// start with an empty override
		var override terra.VersionOverride

		// if an override was set in the file using `--versions-file` flag, use that
		if fromFile, exists := pinOptions.FileOverrides[r.Name()]; exists {
			override = fromFile
		}

		// if global --terra-helmfile-ref was set, add it to our release override
		if pinOptions.Flags.TerraHelmfileRef != "" {
			override.TerraHelmfileRef = pinOptions.Flags.TerraHelmfileRef
		}

		// if global --firecloud-develop-ref was set, add it to our release override
		if pinOptions.Flags.FirecloudDevelopRef != "" {
			override.FirecloudDevelopRef = pinOptions.Flags.FirecloudDevelopRef
		}

		releaseOverrides[r.Name()] = override
	}
	releaseOverridesJson, err := json.Marshal(releaseOverrides)
	if err != nil {
		return nil, err
	}

	log.Debug().Bytes("overrides", releaseOverridesJson).Msgf("Updating release version overrides for %s", bee.Name())

	_, err = b.state.Environments().PinVersions(bee.Name(), releaseOverrides)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Updated version overrides for %s", bee.Name())

	// reload state since we mutated an environment
	if err := b.reloadState(); err != nil {
		return nil, err
	}
	// return a refreshed/updated bee environment object that includes the version overrides
	return b.state.Environments().Get(bee.Name())
}

func (b *bees) UnpinVersions(bee terra.Environment) error {
	wasTerraHelmfileRef := bee.TerraHelmfileRef()
	removed, err := b.state.Environments().UnpinVersions(bee.Name())
	if err != nil {
		return err
	}
	asJson, err := json.Marshal(removed)
	if err != nil {
		return err
	}
	log.Info().Msgf("Removed terra-helmfile version overrides for %s (was: %s)", bee.Name(), wasTerraHelmfileRef)
	log.Info().Bytes("was", asJson).Msgf("Removed all release version overrides for %s", bee.Name())

	return nil
}

func (b *bees) GetBee(name string) (terra.Environment, error) {
	env, err := b.state.Environments().Get(name)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, errors.Errorf("no BEE by the name %q exists", name)
	}
	if env.Lifecycle() != terra.Dynamic {
		return nil, errors.Errorf("environment %s is not a BEE (lifecycle is %s)", name, env.Lifecycle())
	}
	return env, nil
}

func (b *bees) GetTemplate(name string) (terra.Environment, error) {
	template, err := b.state.Environments().Get(name)
	if err != nil {
		return nil, err
	}
	if template != nil && template.Lifecycle() == terra.Template {
		return template, nil
	}

	names, err := b.templateNames()
	if err != nil {
		return nil, err
	}
	return nil, errors.Errorf("no template by the name %q exists, valid templates are: %s", name, strings.Join(names, ", "))
}

func (b *bees) ResetStatefulSets(env terra.Environment) (map[terra.Release]*status.Status, error) {
	var err error

	if err = b.kubectl.ShutDown(env); err != nil {
		return nil, err
	}
	if err = b.kubectl.DeletePVCs(env); err != nil {
		return nil, err
	}

	log.Info().Msgf("Syncing ArgoCD to provision new disks and bring services back up")
	return b.SyncArgoAppsIn(env, func(options *argocd.SyncOptions) {
		options.SyncIfNoDiff = true
	})
}

func (b *bees) Seeder() seed.Seeder {
	return b.seeder
}

func (b *bees) FilterBees(_filter terra.EnvironmentFilter) ([]terra.Environment, error) {
	_filter = filter.Environments().HasLifecycle(terra.Dynamic).And(_filter)
	allEnvs, err := b.state.Environments().All()
	if err != nil {
		return nil, err
	}
	return _filter.Filter(allEnvs), nil
}

func (b *bees) templateNames() ([]string, error) {
	allEnvs, err := b.state.Environments().All()
	if err != nil {
		return nil, err
	}

	templates := filter.Environments().IsTemplate().Filter(allEnvs)

	var names []string
	for _, t := range templates {
		names = append(names, t.Name())
	}
	return names, nil
}

func (b *bees) reloadState() error {
	log.Debug().Msgf("reloading state from Sherlock...")
	state, err := b.stateLoader.Reload()
	if err != nil {
		return err
	}
	b.state = state
	return nil
}
