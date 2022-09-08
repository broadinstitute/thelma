package sherlock

import (
	"fmt"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type environments struct {
	state *state
}

func newEnvironments(s *state) terra.Environments {
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
		return nil, nil
	}
	return env, nil
}

func (e *environments) Exists(name string) (bool, error) {
	_, exists := e.state.environments[name]
	return exists, nil
}

func (e *environments) CreateFromTemplate(name string, template terra.Environment) error {
	return e.CreateHybridFromTemplate(name, template, nil)
}

func (e *environments) CreateHybridFromTemplate(name string, template terra.Environment, fiab terra.Fiab) error {
	exists, err := e.Exists(name)
	if err != nil {
		return fmt.Errorf("error checking for environment name conflict: %v", err)
	}
	if exists {
		return fmt.Errorf("can't create environment: %s: an environment of the same name already exists", template.Name())
	}

	if !template.Lifecycle().IsTemplate() {
		return fmt.Errorf("can't create from template: environment %s is not a template", template.Name())
	}

	panic("TODO")

}

func (e *environments) EnableRelease(environmentName string, releaseName string) error {
	panic("TODO")
}

func (e *environments) DisableRelease(environmentName string, releaseName string) error {
	panic("TODO")
}

func (e *environments) PinVersions(environmentName string, versions map[string]terra.VersionOverride) (map[string]terra.VersionOverride, error) {
	panic("TODO")
}

func (e *environments) PinEnvironmentToTerraHelmfileRef(environmentName string, terraHelmfileRef string) error {
	panic("TODO")
}

func (e *environments) UnpinVersions(environmentName string) (map[string]terra.VersionOverride, error) {
	panic("TODO")
}

func (e *environments) SetBuildNumber(environmentName string, buildNumber int) (int, error) {
	panic("TODO")
}

func (e *environments) UnsetBuildNumber(environmentName string) (int, error) {
	panic("TODO")
}

func (e *environments) Delete(name string) error {
	panic("TODO")
}
