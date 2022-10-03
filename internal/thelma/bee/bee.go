package bee

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/bee/seed"
	slackapi "github.com/broadinstitute/thelma/internal/thelma/clients/slack"
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
	Name         string
	NamePrefix   string
	GenerateName bool
	Template     string
	Owner        string
	Hybrid       bool
	Fiab         struct {
		Name string
		IP   string
	}
	SyncGeneratorOnly bool
	WaitHealthy       bool
	PinOptions        PinOptions
	Seed              bool
	SeedOptions       seed.SeedOptions
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

func NewBees(argocd argocd.ArgoCD, slack slackapi.SlackAPI, stateLoader terra.StateLoader, seeder seed.Seeder, kubectl kubectl.Kubectl) (Bees, error) {
	state, err := stateLoader.Load()
	if err != nil {
		return nil, err
	}

	return &bees{
		argocd:      argocd,
		slack:       slack,
		state:       state,
		stateLoader: stateLoader,
		seeder:      seeder,
		kubectl:     kubectl,
	}, nil
}

// implements Bees interface
type bees struct {
	argocd      argocd.ArgoCD
	slack       slackapi.SlackAPI
	state       terra.State
	stateLoader terra.StateLoader
	seeder      seed.Seeder
	kubectl     kubectl.Kubectl
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
		_options.WaitHealthy = options.WaitHealthy
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

	// Three cases:
	// - Create when user specifies --owner (ALWAYS from Jenkins, b/c user will specify their own email in the Jenkins
	// job options when they create a BEE) --owner is the full email, you care about everything before the @
	// https://fc-jenkins.dsp-techops.broadinstitute.org/view/BEEs/job/fiab-host-create-bee/configure
	// 		Easy! You have the email right here as options.owner
	//		Start here, because you need to be able to handle "what if the email is garbage or doesn't exist?!?"
	// - Create when user doesn't specify --owner (always from local)
	//      User probably won't provide their email
	//		Where to get it from?
	//			- Thelma's understanding of Google authentication--Clients.Google.Terra.GoogleUserInfo() has the email in it
	//				Why not here? Because Sherlock auto-fills the environment owner!!!
	//				In other words, if --owner wasn't provided on command line... and then wasn't sent in the create environment request...
	//				then Sherlock would have done logic to come up with the right owner. If we *also* try to come up with default email...
	//				we are duplicating Sherlock's logic.
	//			- The Environment that comes back from Sherlock when Thelma creates it--it'll ALWAYS have an owner field on it
	//				Current problems:
	//					- Thelma's Environment type doesn't have an Owner field
	//					- Thelma's `err = b.state.Environments().CreateFromTemplate(name, template)` doesn't actually return an Environment :/
	//					  ...but it will once someone gets BEE creation working in this exact function
	//				When you get here... actually just always pay attention to Sherlock, because someone will have wired
	//				--owner into the request to Sherlock in the first place
	// - Delete
	//		Oops, same as above! Because even in Jenkins, we will *always* need to get the email from the owner field
	//		on the existing environment, because user doesn't specify their username when they destroy a BEE, only upon create.
	//		Basically the exact same case as the "Create when user doesn't specify --owner"
	if err = b.sendSlackMsg(options.Owner); err != nil {
		log.Warn().Msgf("Unable to send slack message: %v", err)
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

func (b *bees) sendSlackMsg(owner string) error {
	return b.slack.SendDMMessage(owner)
	//ToDO Handle error messages
}
