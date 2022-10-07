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

func (e *environments) CreateFromTemplate(name string, template terra.Environment) (terra.Environment, error) {
	exists, err := e.Exists(name)
	if err != nil {
		return nil, fmt.Errorf("error checking for environment name conflict: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("can't create environment: %s: an environment of the same name already exists", template.Name())
	}

	if !template.Lifecycle().IsTemplate() {
		return nil, fmt.Errorf("can't create from template: environment %s is not a template", template.Name())
	}

	return buildDynamicEnvironment(template, name, e.state.sherlock)
}

func (e *environments) CreateFromTemplateGenerateName(namePrefix string, template terra.Environment) (terra.Environment, error) {
	if !template.Lifecycle().IsTemplate() {
		return nil, fmt.Errorf("can't create from template: environment %s is not a template", template.Name())
	}
	// sherlock will autogenerate names for dynamic envs so we don't need to specify one
	return buildDynamicEnvironment(template, "", e.state.sherlock)
}

func (e *environments) EnableRelease(environmentName string, releaseName string) error {
	panic("TODO")
}

func (e *environments) DisableRelease(environmentName string, releaseName string) error {
	panic("TODO")
}

// TODO use a real implmentation of this
func (e *environments) PinVersions(environmentName string, versions map[string]terra.VersionOverride) (map[string]terra.VersionOverride, error) {
	// panic("TODO")
	return nil, nil
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
	env, err := e.Get(name)
	if err != nil {
		return err
	}
	_, err = e.state.sherlock.DeleteEnvironments([]terra.Environment{env})
	return err
}

func buildDynamicEnvironment(template terra.Environment, name string, writer terra.StateWriter) (*environment, error) {
	dynamicEnvReleases := make(map[string]*appRelease)

	for _, r := range template.Releases() {
		templateRelease := r.(*appRelease)
		newRelease := &appRelease{
			appVersion: templateRelease.AppVersion(),
			subdomain:  templateRelease.Subdomain(),
			protocol:   templateRelease.Protocol(),
			port:       templateRelease.Port(),
			release: release{
				name:         templateRelease.Name(),
				enabled:      templateRelease.enabled,
				releaseType:  templateRelease.Type(),
				chartVersion: templateRelease.ChartVersion(),
				chartName:    templateRelease.ChartName(),
				repo:         templateRelease.Repo(),
				namespace:    templateRelease.Namespace(),
				cluster:      templateRelease.Cluster(),
				destination:  nil,
				helmfileRef:  template.TerraHelmfileRef(),
			},
		}
		dynamicEnvReleases[newRelease.Name()] = newRelease
	}

	env := &environment{
		defaultCluster:     template.DefaultCluster(),
		releases:           dynamicEnvReleases,
		lifecycle:          terra.Dynamic,
		template:           template.Name(),
		baseDomain:         template.BaseDomain(),
		namePrefixesDomain: template.NamePrefixesDomain(),
		// TODO use a real unique prefix
		uniqueResourcePrefix: "blah",
		destination: destination{
			base:            template.Base(),
			requireSuitable: template.RequireSuitable(),
			destinationType: terra.EnvironmentDestination,
		},
	}
	for _, r := range env.Releases() {
		r.(*appRelease).destination = env
	}

	newEnvs := make([]terra.Environment, 0)
	newEnvs = append(newEnvs, env)
	envNames, err := writer.WriteEnvironments(newEnvs)
	if err != nil {
		return nil, err
	}

	if len(envNames) != 1 {
		return nil, fmt.Errorf("expected only 1 environment to be created but received multiple: %v", envNames)
	}

	env.name = envNames[0]
	return env, nil
}
