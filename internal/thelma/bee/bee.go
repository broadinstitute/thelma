package bee

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/tools/argocd"
	"github.com/rs/zerolog/log"
	"strings"
)

const generatorArgoApp = "terra-bee-generator"

type Bees interface {
	DeleteWith(name string, options DeleteOptions) (terra.Environment, error)
	CreateWith(name string, options CreateOptions) (terra.Environment, error)
	GetTemplate(name string) (terra.Environment, error)
	SyncGeneratorForName(name string) error
	SyncGeneratorFor(env terra.Environment) error
	SyncArgoAppsFor(env terra.Environment, options ...argocd.SyncOption) error
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
	GeneratorOnly bool
	WaitHealthy   bool
}

type VersionOptions struct {
	AppVersion          string
	ChartVersion        string
	TerraHelmfileRef    string
	FirecloudDevelopRef string
}

func NewBees(argocd argocd.ArgoCD, stateLoader terra.StateLoader) (Bees, error) {
	state, err := stateLoader.Load()
	if err != nil {
		return nil, err
	}

	return &bees{
		argocd:      argocd,
		state:       state,
		stateLoader: stateLoader,
	}, nil
}

// implements Bees interface
type bees struct {
	argocd      argocd.ArgoCD
	state       terra.State
	stateLoader terra.StateLoader
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

	if err = b.SyncGenerator(); err != nil {
		return env, err
	}
	if options.GeneratorOnly {
		log.Warn().Msgf("Won't sync Argo apps for %s", env.Name())
		return env, nil
	}
	log.Info().Msgf("Syncing all Argo apps in environment %s", env.Name())
	err = b.SyncArgoAppsFor(env, func(_options *argocd.SyncOptions) {
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

	if err = b.state.Environments().Delete(env.Name()); err != nil {
		return nil, err
	}

	log.Info().Msgf("Deleted environment %s from state", name)
	log.Info().Msgf("Deleting Argo apps for %s", name)
	if err = b.SyncGenerator(); err != nil {
		return env, err
	}

	log.Info().Msgf("Deleting Argo project for %s", name)
	if err = b.SyncGenerator(); err != nil {
		return env, err
	}

	return env, nil
}

func (b *bees) SyncGeneratorForName(name string) error {
	env, err := b.state.Environments().Get(name)
	if err != nil {
		return err
	}
	if env == nil {
		return fmt.Errorf("no such bee: %s", name)
	}
	return b.SyncGeneratorFor(env)
}

func (b *bees) SyncGeneratorFor(env terra.Environment) error {
	log.Info().Msgf("Syncing %s for %s", generatorArgoApp, env.Name())
	return b.syncGenerator(func(options *argocd.SyncOptions) {
		options.OnlyLabels = argocd.EnvironmentSelector(env)
	})
}

func (b *bees) SyncArgoAppsFor(env terra.Environment, options ...argocd.SyncOption) error {
	releases, err := b.state.Releases().Filter(filter.Releases().BelongsToEnvironment(env))
	if err != nil {
		return err
	}
	return b.argocd.SyncReleases(releases, 15, options...)
}

func (b *bees) SyncGenerator() error {
	log.Info().Msgf("Syncing %s", generatorArgoApp)
	return b.syncGenerator()
}

func (b *bees) syncGenerator(options ...argocd.SyncOption) error {
	options = append(options, func(options *argocd.SyncOptions) {
		// never wait for generator to be healthy -- it gets a lot of traffic that can cause it to go out of sync
		options.WaitHealthy = false
	})
	return b.argocd.SyncApp(generatorArgoApp, options...)
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
