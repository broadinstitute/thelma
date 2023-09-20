package sherlock

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type environments struct {
	state *state
}

func newEnvironmentsView(s *state) terra.Environments {
	return &environments{
		state: s,
	}
}

func (e *environments) All() ([]terra.Environment, error) {
	var result []terra.Environment
	for _, environment := range e.state.environments {
		result = append(result, environment)
	}

	return result, nil
}

func (e *environments) Filter(filter terra.EnvironmentFilter) ([]terra.Environment, error) {
	all, err := e.All()
	if err != nil {
		return nil, err
	}

	return filter.Filter(all), nil
}

func (e *environments) Get(name string) (terra.Environment, error) {
	env, exists := e.state.environments[name]
	if !exists {
		return nil, errors.Errorf("environment %q does not exist", name)
	}
	return env, nil
}

func (e *environments) Exists(name string) (bool, error) {
	_, exists := e.state.environments[name]
	return exists, nil
}

func (e *environments) CreateFromTemplate(template terra.Environment, options terra.CreateOptions) (string, error) {
	return e.state.sherlock.CreateEnvironmentFromTemplate(template.Name(), options)
}

func (e *environments) EnableRelease(environmentName string, releaseName string) error {
	environment, err := e.Get(environmentName)
	if err != nil {
		return err
	}
	if environment.Lifecycle() != terra.Dynamic {
		return errors.Errorf("enabling releases is only supported for dynamic environments")
	}
	return e.state.sherlock.EnableRelease(environment, releaseName)
}

func (e *environments) DisableRelease(environmentName string, releaseName string) error {
	environment, err := e.Get(environmentName)
	if err != nil {
		return err
	}
	if environment.Lifecycle() != terra.Dynamic {
		return errors.Errorf("disabling releases is only supported in dynamic environments")
	}
	return e.state.sherlock.DisableRelease(environmentName, releaseName)
}

func (e *environments) PinVersions(environmentName string, versions map[string]terra.VersionOverride) (map[string]terra.VersionOverride, error) {
	return versions, e.state.sherlock.PinEnvironmentVersions(environmentName, versions)
}

func (e *environments) PinEnvironmentToTerraHelmfileRef(environmentName string, terraHelmfileRef string) error {
	environment, err := e.Get(environmentName)
	if err != nil {
		return err
	}
	return e.state.sherlock.SetTerraHelmfileRefForEntireEnvironment(environment, terraHelmfileRef)
}

func (e *environments) UnpinVersions(environmentName string) (map[string]terra.VersionOverride, error) {
	environment, err := e.Get(environmentName)
	if err != nil {
		return nil, err
	}
	log.Warn().Msg("sherlock state provider does not directly support unpinning, so the environment's git refs will be reset and it will be pinned to the current state of dev")
	log.Debug().Msg("note that because sherlock does not store overrides, thelma will report that zero overrides were lifted")
	return nil, e.state.sherlock.ResetEnvironmentAndPinToDev(environment)
}

func (e *environments) Delete(name string) error {
	env, err := e.Get(name)
	if err != nil {
		return err
	}
	_, err = e.state.sherlock.DeleteEnvironments([]terra.Environment{env})
	return err
}

func (e *environments) SetOffline(name string, offline bool) error {
	return e.state.sherlock.SetEnvironmentOffline(name, offline)
}
