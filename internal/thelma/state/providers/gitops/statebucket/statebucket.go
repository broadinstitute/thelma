package statebucket

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"sort"
	"time"
)

const configKey = "statebucket"

// bump this whenever backwards-incompatible schema changes are made. This way any clients that attempt to update
// the state with the old schema will return an error.
const schemaVersion = 1

// StateFile represents the structure of the statefile
type StateFile struct {
	SchemaVersion int32                         `json:"schemaVersion"`
	Environments  map[string]DynamicEnvironment `json:"environments"`
}

type statebucketConfig struct {
	// Name of the GCS bucket where state file is kept
	Name string `default:"thelma-state"`
	// Object name of the object in the bucket where state is kept
	Object string `default:"state.json"`
	// Lock settings for the state lock used to prevent concurrent updates from stomping on each other
	Lock struct {
		Object       string        `default:".update.lk"`
		MaxWait      time.Duration `default:"30s"`
		ExpiresAfter time.Duration `default:"5m"`
	}
}

// StateBucket is for tracking state for dynamic environments. (Stored in a GCS bucket)
type StateBucket interface {
	// Environments returns the list of all environments in the state file
	Environments() ([]DynamicEnvironment, error)
	// Add adds a new environment to the state file
	Add(environment DynamicEnvironment) error
	// EnableRelease enables the given release in the target environment
	EnableRelease(environmentName string, releaseName string) error
	// DisableRelease disables the given release in the target environment
	DisableRelease(environmentName string, releaseName string) error
	// PinVersions can be used to update the environment's map of version overrides
	PinVersions(environmentName string, versions map[string]terra.VersionOverride) (map[string]terra.VersionOverride, error)
	// UnpinVersions can be used to remove the environment's map of version overrides
	UnpinVersions(environmentName string) (map[string]terra.VersionOverride, error)
	// PinEnvironmentToTerraHelmfileRef pins an entire environment to a specific terra-helmfile ref
	PinEnvironmentToTerraHelmfileRef(environmentName string, terraHelmfileRef string) error
	// Delete will delete an environment from the state file
	Delete(environmentName string) error
	// initialize will overwrite existing state with a new empty state file
	initialize() error
}

// New returns a new statebucket
func New(thelmaConfig config.Config, googleClients google.Clients) (StateBucket, error) {
	cfg, err := loadConfig(thelmaConfig)
	if err != nil {
		return nil, err
	}

	_bucket, err := googleClients.Bucket(cfg.Name)
	if err != nil {
		return nil, fmt.Errorf("error initializing state bucket %s: %v", cfg.Name, err)
	}

	return newWithBucket(_bucket, cfg), nil
}

// NewFake (FOR USE IN TESTS ONLY) returns a new fake statebucket, backed by local filesystem instead of a GCS bucket
func NewFake(dir string) (StateBucket, error) {
	return &statebucket{
		writer: newSchemaVerifier(schemaVersion, newFileWriter(dir, "state.json")),
	}, nil
}

// package-private constructor, used in testing
func newWithBucket(_bucket bucket.Bucket, cfg statebucketConfig) *statebucket {
	return &statebucket{
		writer: newSchemaVerifier(schemaVersion, newBucketWriter(_bucket, cfg)),
	}
}

func loadConfig(thelmaConfig config.Config) (statebucketConfig, error) {
	var cfg statebucketConfig
	err := thelmaConfig.Unmarshal(configKey, &cfg)
	return cfg, err
}

type statebucket struct {
	writer writer
}

func (s *statebucket) Environments() ([]DynamicEnvironment, error) {
	state, err := s.writer.read()
	if err != nil {
		return nil, err
	}

	var result []DynamicEnvironment

	if state.Environments == nil {
		return result, nil
	}

	for _, env := range state.Environments {
		result = append(result, env)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (s *statebucket) Add(environment DynamicEnvironment) error {
	return s.writer.update(func(state StateFile) (StateFile, error) {
		if state.Environments == nil {
			state.Environments = make(map[string]DynamicEnvironment)
		}

		_, exists := state.Environments[environment.Name]
		if exists {
			return StateFile{}, fmt.Errorf("can't add new environment %s, an environment by that name already exists", environment.Name)
		}
		state.Environments[environment.Name] = environment
		return state, nil
	})
}

func (s *statebucket) EnableRelease(environmentName string, releaseName string) error {
	return s.updateEnvironment(environmentName, func(e *DynamicEnvironment) {
		e.setOverride(releaseName, func(override *Override) {
			override.Enable()
		})
	})
}

func (s *statebucket) DisableRelease(environmentName string, releaseName string) error {
	return s.updateEnvironment(environmentName, func(e *DynamicEnvironment) {
		e.setOverride(releaseName, func(override *Override) {
			override.Disable()
		})
	})
}

func (s *statebucket) PinVersions(environmentName string, versions map[string]terra.VersionOverride) (map[string]terra.VersionOverride, error) {
	result := make(map[string]terra.VersionOverride)

	err := s.updateEnvironment(environmentName, func(e *DynamicEnvironment) {
		for releaseName, v := range versions {
			e.setOverride(releaseName, func(override *Override) {
				override.PinVersions(v)
			})
		}

		for releaseName, override := range e.Overrides {
			result[releaseName] = override.Versions
		}
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *statebucket) PinEnvironmentToTerraHelmfileRef(environmentName string, terraHelmfileRef string) error {
	return s.updateEnvironment(environmentName, func(environment *DynamicEnvironment) {
		environment.TerraHelmfileRef = terraHelmfileRef
	})
}

func (s *statebucket) UnpinVersions(environmentName string) (map[string]terra.VersionOverride, error) {
	result := make(map[string]terra.VersionOverride)

	err := s.updateEnvironment(environmentName, func(e *DynamicEnvironment) {
		var deletions []string

		e.TerraHelmfileRef = ""

		for releaseName, override := range e.Overrides {
			result[releaseName] = override.Versions
			override.UnpinVersions()
			if !override.HasEnableOverride() {
				// no version or enable override, so we should delete the key
				deletions = append(deletions, releaseName)
			}
		}

		for _, releaseName := range deletions {
			delete(e.Overrides, releaseName)
		}
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *statebucket) Delete(environmentName string) error {
	return s.writer.update(func(state StateFile) (StateFile, error) {
		_, exists := state.Environments[environmentName]
		if !exists {
			return StateFile{}, fmt.Errorf("can't delete environment %s, it does not exist in the state file", environmentName)
		}
		delete(state.Environments, environmentName)
		return state, nil
	})
}

func (s *statebucket) updateEnvironment(environmentName string, updateFn func(environment *DynamicEnvironment)) error {
	return s.writer.update(func(state StateFile) (StateFile, error) {
		environment, exists := state.Environments[environmentName]
		if !exists {
			return StateFile{}, fmt.Errorf("can't update environment %s, it does not exist in the state file", environmentName)
		}
		updateFn(&environment)
		state.Environments[environmentName] = environment
		return state, nil
	})
}

// populate a new empty statefile in the bucket
func (s *statebucket) initialize() error {
	return s.writer.write(StateFile{SchemaVersion: schemaVersion})
}
