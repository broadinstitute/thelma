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

// CreateFromTemplate doesn't actually need to call buildDynamicEnvironment, because it only needs to return the name.
// The required side effect is to record enough state that the name can be used to (re)construct the dynamic environment,
// so what's needed here is pretty minimal--the heavy lifting will be done when Thelma reloads the state and grabs
// this environment then.
// This slightly-roundabout mechanism is here because it makes a ton of sense for Sherlock, which does all its logic
// in the side-effect step here and has a very thin "get" step later after Thelma reloads.
func (e *environments) CreateFromTemplate(template terra.Environment, options terra.CreateOptions) (string, error) {
	exists, err := e.Exists(options.Name)
	if options.Name == "" {
		return "", fmt.Errorf("generating names not supported")
	}
	if err != nil {
		return "", fmt.Errorf("error checking for environment name conflict: %v", err)
	}
	if exists {
		return "", fmt.Errorf("can't create environment %s: an environment by that name already exists", options.Name)
	}

	if !template.Lifecycle().IsTemplate() {
		return "", fmt.Errorf("can't create from template: environment %s is not a template", template.Name())
	}

	var env statebucket.DynamicEnvironment
	env.Name = options.Name
	env.Template = template.Name()

	env, err = e.state.statebucket.Add(env)
	if err != nil {
		return "", err
	}

	return env.Name, nil
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

func (e *environments) SetOffline(_ string, _ bool) error {
	return nil
}
