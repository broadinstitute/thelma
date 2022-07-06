package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statebucket"
)

func newEnvironments(g *state) terra.Environments {
	return &environments{
		state: g,
	}
}

type environments struct {
	state *state
}

func (e *environments) All() ([]terra.Environment, error) {
	var result []terra.Environment

	for _, env := range e.state.environments {
		result = append(result, env)
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
		return fmt.Errorf("can't create environment %s: an environment by that name already exists", name)
	}

	if !template.Lifecycle().IsTemplate() {
		return fmt.Errorf("can't create from template: environment %s is not a template", template.Name())
	}

	var env statebucket.DynamicEnvironment
	env.Name = name
	env.Template = template.Name()

	if fiab != nil {
		env.Hybrid = true
		env.Fiab = statebucket.Fiab{
			Name: fiab.Name(),
			IP:   fiab.IP(),
		}
	}

	return e.state.statebucket.Add(env)
}

func (e *environments) EnableRelease(environmentName string, releaseName string) error {
	return e.state.statebucket.EnableRelease(environmentName, releaseName)
}

func (e *environments) DisableRelease(environmentName string, releaseName string) error {
	return e.state.statebucket.DisableRelease(environmentName, releaseName)
}

func (e *environments) PinVersions(environmentName string, versions map[string]terra.VersionOverride) (map[string]terra.VersionOverride, error) {
	return e.state.statebucket.PinVersions(environmentName, versions)
}

func (e *environments) PinEnvironmentToTerraHelmfileRef(environmentName string, terraHelmfileRef string) error {
	return e.state.statebucket.PinEnvironmentToTerraHelmfileRef(environmentName, terraHelmfileRef)
}

func (e *environments) UnpinVersions(environmentName string) (map[string]terra.VersionOverride, error) {
	return e.state.statebucket.UnpinVersions(environmentName)
}

func (e *environments) SetBuildNumber(environmentName string, buildNumber int) (int, error) {
	return e.state.statebucket.SetBuildNumber(environmentName, buildNumber)
}

func (e *environments) UnsetBuildNumber(environmentName string) (int, error) {
	return e.state.statebucket.UnsetBuildNumber(environmentName)
}

func (e *environments) Delete(name string) error {
	return e.state.statebucket.Delete(name)
}
