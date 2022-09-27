package bee

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/tools/kubectl"
	"github.com/rs/zerolog/log"
	"strings"
)

const generatorArgoApp = "terra-bee-generator"

type Bees interface {
	DeleteWith(name string, options DeleteOptions) (terra.Environment, error)
	CreateWith(name string, options CreateOptions) (terra.Environment, error)
	GetTemplate(templateName string) (terra.Environment, error)
	RefreshBeeGenerator() error
	SyncEnvironmentGenerator(env terra.Environment) error
	SyncArgoAppsIn(env terra.Environment, options ...argocd.SyncOption) error
	ResetStatefulSets(env terra.Environment) error
}

type DeleteOptions struct {
	IgnoreMissing bool
}

type CreateOptions struct {
	Template string
	Hybrid   bool
	Fiab     struct {
		Name string
		IP   string
	}
	SyncGeneratorOnly bool
	WaitHealthy       bool
	TerraHelmfileRef  string
}

func NewBees(argocd argocd.ArgoCD, stateLoader terra.StateLoader, kubectl kubectl.Kubectl) (Bees, error) {
	state, err := stateLoader.Load()
	if err != nil {
		return nil, err
	}

	return &bees{
		argocd:      argocd,
		state:       state,
		stateLoader: stateLoader,
		kubectl:     kubectl,
	}, nil
}

// implements Bees interface
type bees struct {
	argocd      argocd.ArgoCD
	state       terra.State
	stateLoader terra.StateLoader
	kubectl     kubectl.Kubectl
}

func (b *bees) CreateWith(name string, options CreateOptions) (terra.Environment, error) {
	template, err := b.GetTemplate(options.Template)

	if err != nil {
		return nil, err
	}

	if options.Hybrid {
		err = b.state.Environments().CreateHybridFromTemplate(name, template, terra.NewFiab(options.Fiab.Name, options.Fiab.IP))
	} else {
		err = b.state.Environments().CreateFromTemplate(name, template)
	}
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Created new environment %s", name)

	if options.TerraHelmfileRef != "" {
		log.Info().Msgf("Pinning %s to terra-helmfile ref: %s", name, options.TerraHelmfileRef)
		if err = b.state.Environments().PinEnvironmentToTerraHelmfileRef(name, options.TerraHelmfileRef); err != nil {
			return nil, err
		}
	}

	if err = b.reloadState(); err != nil {
		return nil, err
	}
	env, err := b.state.Environments().Get(name)
	if err != nil {
		return nil, err
	}
	if env == nil {
		// don't think this could ever happen, but let's provide a useful error anyway
		return nil, fmt.Errorf("error creating environment %q: missing from state after creation", name)
	}

	log.Info().Msgf("Creating environment namespace for %s", env.Name())
	err = b.kubectl.CreateNamespace(env)
	if err != nil {
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
		// (the deployments are being created for the first time, so they'll definitely pick up changes)
		_options.SkipLegacyConfigsRestart = true
		_options.WaitHealthy = options.WaitHealthy
	})
	return env, err
}

func (b *bees) DeleteWith(name string, options DeleteOptions) (terra.Environment, error) {
	env, err := b.state.Environments().Get(name)
	if err != nil {
		return nil, err
	}

	if env == nil {
		if options.IgnoreMissing {
			log.Warn().Msgf("Could not delete %s, no BEE by that name exists", name)
			return nil, nil
		} else {
			return nil, fmt.Errorf("delete %s failed: no BEE by that name exists", name)
		}
	}

	log.Info().Msgf("Deleting environment namespace")
	if err = b.kubectl.DeleteNamespace(env); err != nil {
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
