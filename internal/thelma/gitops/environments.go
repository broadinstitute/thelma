package gitops

import "github.com/broadinstitute/thelma/internal/thelma/terra"

type environments struct {
	state *gitops
}

func newEnvironments(g *gitops) terra.Environments {
	return &environments{
		state: g,
	}
}

func (e *environments) All() ([]terra.Environment, error) {
	var result []terra.Environment

	for _, env := range e.state.environments {
		result = append(result, env)
	}

	return result, nil
}

func (e *environments) Get(name string) (terra.Environment, error) {
	return e.state.environments[name], nil
}

func (e *environments) Exists(name string) (bool, error) {
	_, exists := e.state.environments[name]
	return exists, nil
}

func (e *environments) CreateFromTemplate(name string, template terra.Environment) error {
	//TODO implement me
	panic("implement me")
}

func (e *environments) PinVersions(name string, versions map[string]string) error {
	//TODO implement me
	panic("implement me")
}

func (e *environments) UnpinVersions(name string) error {
	//TODO implement me
	panic("implement me")
}

func (e *environments) Delete(name string) error {
	//TODO implement me
	panic("implement me")
}
