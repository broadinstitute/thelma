package seed

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/clients"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/tools/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
)

type commonOptions struct {
	Force   bool
	NoSteps bool
}

type SeedOptions struct {
	Step1CreateElasticsearch bool
	Step2RegisterSaProfiles  bool
	Step3AddSaSamPermissions bool
	Step4RegisterTestUsers   bool
	Step5CreateAgora         bool
	Step6ExtraUser           []string
	RegisterSelfShortcut     bool
	commonOptions
}

type UnseedOptions struct {
	Step1UnregisterAllUsers bool
	commonOptions
}

type Seeder interface {
	Seed(env terra.Environment, seedOptions SeedOptions) error
	Unseed(env terra.Environment, unseedOptions UnseedOptions) error
}

type seeder struct {
	config        config.Config
	kubectl       kubectl.Kubectl
	clientFactory clients.Clients
	shellRunner   shell.Runner
}

func New(kubectl kubectl.Kubectl, clientFactory clients.Clients, thelmaConfig config.Config, shellRunner shell.Runner) Seeder {
	return &seeder{
		kubectl:       kubectl,
		clientFactory: clientFactory,
		config:        thelmaConfig,
		shellRunner:   shellRunner,
	}
}

func (s *seeder) Seed(bee terra.Environment, seedOptions SeedOptions) error {
	appReleases := getAppReleases(bee)

	if seedOptions.Step1CreateElasticsearch {
		if err := seedOptions.handleErrorWithForce(s.seedStep1CreateElasticsearch(appReleases, seedOptions)); err != nil {
			return err
		}
	}
	if seedOptions.Step2RegisterSaProfiles {
		if err := seedOptions.handleErrorWithForce(s.seedStep2RegisterSaProfiles(appReleases, seedOptions)); err != nil {
			return err
		}
	}
	if seedOptions.Step3AddSaSamPermissions {
		if err := seedOptions.handleErrorWithForce(s.seedStep3AddSaSamPermissions(appReleases, seedOptions)); err != nil {
			return err
		}
	}
	if seedOptions.Step4RegisterTestUsers {
		if err := seedOptions.handleErrorWithForce(s.seedStep4RegisterTestUsers(appReleases, seedOptions)); err != nil {
			return err
		}
	}
	if seedOptions.Step5CreateAgora {
		if err := seedOptions.handleErrorWithForce(s.seedStep5CreateAgora(appReleases, seedOptions)); err != nil {
			return err
		}
	}
	if len(seedOptions.Step6ExtraUser) > 0 {
		if err := seedOptions.handleErrorWithForce(s.seedStep6ExtraUser(appReleases, seedOptions)); err != nil {
			return err
		}
	}

	return nil
}

func (s *seeder) Unseed(bee terra.Environment, unseedOptions UnseedOptions) error {
	appReleases := getAppReleases(bee)

	if unseedOptions.Step1UnregisterAllUsers {
		if err := unseedOptions.handleErrorWithForce(s.unseedStep1UnregisterAllUsers(appReleases, unseedOptions)); err != nil {
			return err
		}
	}

	return nil
}

func (o commonOptions) handleErrorWithForce(err error) error {
	if err != nil && o.Force {
		log.Warn().Msgf("%v", err.Error())
		log.Warn().Msgf("Continuing despite above error (--force seeding option enabled)")
		return nil
	} else {
		return err
	}
}

func getAppReleases(bee terra.Environment) map[string]terra.AppRelease {
	appReleases := make(map[string]terra.AppRelease)
	for _, release := range bee.Releases() {
		if release.IsAppRelease() {
			appRelease, wasAppRelease := release.(terra.AppRelease)
			if wasAppRelease {
				appReleases[appRelease.Name()] = appRelease
			} else {
				log.Warn().Msgf("%s was an App Release but failed to type-assert", release.Name())
			}
		}
	}
	return appReleases
}
