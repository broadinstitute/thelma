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

func (e *environments) CreateFromTemplate(name string, template terra.Environment) (terra.Environment, error) {
	exists, err := e.Exists(name)
	if err != nil {
		return nil, fmt.Errorf("error checking for environment name conflict: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("can't create environment %s: an environment by that name already exists", name)
	}

	if !template.Lifecycle().IsTemplate() {
		return nil, fmt.Errorf("can't create from template: environment %s is not a template", template.Name())
	}

	var env statebucket.DynamicEnvironment
	env.Name = name
	env.Template = template.Name()

	env, err = e.state.statebucket.Add(env)
	if err != nil {
		return nil, err
	}

	return buildDynamicEnvironment(template, env), nil
}

func (e *environments) CreateFromTemplateGenerateName(namePrefix string, template terra.Environment) (terra.Environment, error) {
	if !template.Lifecycle().IsTemplate() {
		return nil, fmt.Errorf("can't create from template: environment %s is not a template", template.Name())
	}
	var env statebucket.DynamicEnvironment
	env.Template = template.Name()
	env, err := e.state.statebucket.AddGenerateName(namePrefix, env)
	if err != nil {
		return nil, err
	}
	return buildDynamicEnvironment(template, env), nil
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

func (e *environments) Delete(name string) error {
	return e.state.statebucket.Delete(name)
}
