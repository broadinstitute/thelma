package bee

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/bee/cleanup"
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/bee/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/tools/kubectl"
	"github.com/rs/zerolog/log"
)

const generatorArgoApp = "terra-bee-generator"

type Bees interface {
	DeleteWith(name string, options DeleteOptions) (terra.Environment, error)
	CreateWith(options CreateOptions) (terra.Environment, error)
	GetBee(name string) (terra.Environment, error)
	GetTemplate(templateName string) (terra.Environment, error)
	Seeder() seed.Seeder
	FilterBees(filter terra.EnvironmentFilter) ([]terra.Environment, error)
	PinVersions(bee terra.Environment, overrides PinOptions) error
	UnpinVersions(bee terra.Environment) error
	SyncEnvironmentGenerator(env terra.Environment) error
	SyncArgoAppsIn(env terra.Environment, options ...argocd.SyncOption) error
	ResetStatefulSets(env terra.Environment) error
	RefreshBeeGenerator() error
}

type DeleteOptions struct {
	Unseed bool
}

type CreateOptions struct {
	Name                      string
	NamePrefix                string
	GenerateName              bool
	Template                  string
	SyncGeneratorOnly         bool
	WaitHealthy               bool
	WaitHealthyTimeoutSeconds int
	PinOptions                PinOptions
	Seed                      bool
	SeedOptions               seed.SeedOptions
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

func NewBees(argocd argocd.ArgoCD, stateLoader terra.StateLoader, seeder seed.Seeder, cleanup cleanup.Cleanup, kubectl kubectl.Kubectl) (Bees, error) {
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
}

func (b *bees) CreateWith(options CreateOptions) (terra.Environment, error) {
	template, err := b.GetTemplate(options.Template)

	if err != nil {
		return nil, err
	}

	var env terra.Environment
	if options.GenerateName {
		env, err = b.state.Environments().CreateFromTemplateGenerateName(options.NamePrefix, template)
	} else {
		env, err = b.state.Environments().CreateFromTemplate(options.Name, template)
	}
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Created new environment %s", env.Name())

	// Load environment from state file
	if err = b.reloadState(); err != nil {
		return nil, err
	}
	env, err = b.state.Environments().Get(env.Name())
	if err != nil {
		return nil, err
	}
	if env == nil {
		// don't think this could ever happen, but let's provide a useful error anyway
		return nil, fmt.Errorf("error creating environment %q: missing from state after creation", env.Name())
	}

	err = b.kubectl.CreateNamespace(env)
	if err != nil {
		return nil, err
	}

	if err = b.PinVersions(env, options.PinOptions); err != nil {
		return nil, err
	}

	if err = b.RefreshBeeGenerator(); err != nil {
		return env, err
	}

	if err = b.argocd.WaitExist(argocd.GeneratorName(env)); err != nil {
		return nil, err
	}
	if err = b.SyncEnvironmentGenerator(env); err != nil {
		return env, err
	}
	if options.SyncGeneratorOnly {
		log.Warn().Msgf("Won't sync Argo apps for %s", env.Name())
		return env, nil
	}

	log.Info().Msgf("Syncing all Argo apps in environment %s", env.Name())
	err = b.SyncArgoAppsIn(env, func(_options *argocd.SyncOptions) {
		// No need to do a legacy configs restart the first time we create a BEE
		// (the deployments are being created for the first time)
		_options.SyncIfNoDiff = true
		_options.SkipLegacyConfigsRestart = true
		_options.WaitHealthy.Enabled = options.WaitHealthy
		_options.WaitHealthy.TimeoutSeconds = options.WaitHealthyTimeoutSeconds
	})
	if err != nil {
		return env, err
	}

	if options.Seed {
		log.Info().Msgf("Seeding BEE with test data")
		if err = b.seeder.Seed(env, options.SeedOptions); err != nil {
			return env, err
		}
	}

	return env, err
}

func (b *bees) DeleteWith(name string, options DeleteOptions) (terra.Environment, error) {
	env, err := b.GetBee(name)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	if err = b.cleanup.Cleanup(env); err != nil {
		return nil, err
	}

	if err = b.state.Environments().Delete(env.Name()); err != nil {
		return nil, err
	}

	log.Info().Msgf("Deleted environment %s from state", name)

	log.Info().Msgf("Deleting Argo apps for %s", name)
	if err = b.RefreshBeeGenerator(); err != nil {
		return env, err
	}

	return env, nil
}

func (b *bees) SyncEnvironmentGenerator(env terra.Environment) error {
	appName := argocd.GeneratorName(env)
	log.Info().Msgf("Syncing generator %s for %s", appName, env.Name())
	_, err := b.argocd.SyncApp(appName)
	return err
}

func (b *bees) SyncArgoAppsIn(env terra.Environment, options ...argocd.SyncOption) error {
	releases, err := b.state.Releases().Filter(filter.Releases().BelongsToEnvironment(env))
	if err != nil {
		return err
	}
	return b.argocd.SyncReleases(releases, len(releases), options...)
}

func (b *bees) RefreshBeeGenerator() error {
	log.Info().Msgf("Refreshing %s", generatorArgoApp)
	// workaround for a bug in ArgoCD:
	//   https://github.com/argoproj/argo-cd/issues/4505#issuecomment-880271371
	// We perform a hard refresh with autosync
	return b.argocd.HardRefresh(generatorArgoApp)
}

func (b *bees) PinVersions(bee terra.Environment, pinOptions PinOptions) error {
	// pin global terra-helmfile ref, if one is specified
	if pinOptions.Flags.TerraHelmfileRef != "" {
		was := bee.TerraHelmfileRef()
		if err := b.state.Environments().PinEnvironmentToTerraHelmfileRef(bee.Name(), pinOptions.Flags.TerraHelmfileRef); err != nil {
			return err
		}
		log.Info().Msgf("Set terra-helmfile ref to %s for %s (was: %s)", bee.Name(), pinOptions.Flags.TerraHelmfileRef, was)
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
		return err
	}

	log.Debug().Bytes("overrides", releaseOverridesJson).Msgf("Updating release version overrides for %s", bee.Name())

	_, err = b.state.Environments().PinVersions(bee.Name(), releaseOverrides)
	if err != nil {
		return err
	}

	log.Info().Msgf("Updated version overrides for %s", bee.Name())
	return nil
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
		return nil, fmt.Errorf("no BEE by the name %q exists", name)
	}
	if env.Lifecycle() != terra.Dynamic {
		return nil, fmt.Errorf("environment %s is not a BEE (lifecycle is %s)", name, env.Lifecycle())
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
	return nil, fmt.Errorf("no template by the name %q exists, valid templates are: %s", name, strings.Join(names, ", "))
}

func (b *bees) ResetStatefulSets(env terra.Environment) error {
	var err error

	if err = b.kubectl.ShutDown(env); err != nil {
		return err
	}
	if err = b.kubectl.DeletePVCs(env); err != nil {
		return err
	}

	log.Info().Msgf("Syncing ArgoCD to provision new disks and bring services back up")
	if err = b.SyncArgoAppsIn(env, func(options *argocd.SyncOptions) {
		options.SyncIfNoDiff = true
	}); err != nil {
		return err
	}

	return nil
}

func (b *bees) Seeder() seed.Seeder {
	return b.seeder
}

func (b *bees) FilterBees(_filter terra.EnvironmentFilter) ([]terra.Environment, error) {
	_filter = filter.Environments().HasLifecycle(terra.Dynamic).And(_filter)
	return b.state.Environments().Filter(_filter)
}

func (b *bees) templateNames() ([]string, error) {
	templates, err := b.state.Environments().Filter(filter.Environments().HasLifecycle(terra.Template))
	if err != nil {
		return nil, err
	}

	var names []string
	for _, t := range templates {
		names = append(names, t.Name())
	}
	return names, nil
}

func (b *bees) reloadState() error {
	state, err := b.stateLoader.Load()
	if err != nil {
		return err
	}
	b.state = state
	return nil
}
