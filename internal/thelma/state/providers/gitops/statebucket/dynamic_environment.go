package statebucket

import (
	"encoding/json"
)

// DynamicEnvironment is a struct representing a dynamic environment in the state file
type DynamicEnvironment struct {
	Name                 string               `json:"name"`
	Template             string               `json:"template"`
	Overrides            map[string]*Override `json:"overrides"`
	TerraHelmfileRef     string               `json:"terraHelmfileRef"`
	UniqueResourcePrefix string               `json:"uniqueResourcePrefix"`
}

// setOverride can be used to update the override for a given release
func (e *DynamicEnvironment) setOverride(releaseName string, setFn func(*Override)) {
	o, exists := e.Overrides[releaseName]
	if !exists {
		o = &Override{}
	}
	setFn(o)
	e.Overrides[releaseName] = o
}

// Note: DynamicEnvironment has custom JSON marshallers/unmarshallers to replace null overrides maps with empty
func (e DynamicEnvironment) MarshalJSON() ([]byte, error) {
	type alias DynamicEnvironment
	aux := struct {
		alias
	}{
		alias: (alias)(e),
	}

	// initialize Overrides with empty map if nil
	if aux.Overrides == nil {
		aux.Overrides = make(map[string]*Override)
	}

	// Temporary hack to populate URP on all environments
	if aux.UniqueResourcePrefix == "" {
		aux.UniqueResourcePrefix = backwardsCompatibleResourcePrefix(e.Name)
	}

	return json.Marshal(aux)
}

func (e *DynamicEnvironment) UnmarshalJSON(data []byte) error {
	type alias DynamicEnvironment
	aux := &struct {
		*alias
	}{
		alias: (*alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// initialize Overrides with empty map if nil
	if e.Overrides == nil {
		e.Overrides = make(map[string]*Override)
	}

	// Temporary hack to populate URP on all environments
	if e.UniqueResourcePrefix == "" {
		e.UniqueResourcePrefix = backwardsCompatibleResourcePrefix(e.Name)
	}

	return nil
}
