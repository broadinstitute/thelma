package statebucket

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/bucket"
	"sort"
	"time"
)

const configKey = "statebucket"

type statebucketConfig struct {
	// Name of the GCS bucket
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

// Fiab DEPRECATED struct for representing a Fiab in state file
type Fiab struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

// DynamicEnvironment is a struct for representing a dynamic environment in the state file
type DynamicEnvironment struct {
	Name        string            `json:"name"`
	Template    string            `json:"template"`
	VersionPins map[string]string `json:"versionPins"`
	Hybrid      bool              `json:"hybrid"` // Deprecated / temporary (while we run bees in hybrid mode)
	Fiab        Fiab              `json:"fiab"`   // Deprecated / temporary (while we run bees in hybrid mode)
}

type StateFile struct {
	Environments map[string]DynamicEnvironment `json:"environments"`
}

// StateBucket is for track state for dynamic environments. (Stored in a GCS bucket)
type StateBucket interface {
	// Environments returns the list of all environments in the state file
	Environments() ([]DynamicEnvironment, error)
	// Add adds a new environment to the state file
	Add(environment DynamicEnvironment) error
	// PinVersions will update the environment's map of version pins in a merging fashion.
	// For example, if the existing pins are {"A":v100", "B":"v200"}, and PinVersions is called
	// with {"A":"v123"}, the new set of pins will be {"A":"v123", "B":"v200"}. UnpinVersions
	// can be used to remove all version pins for the environment.
	PinVersions(environmentName string, versionPins map[string]string) error
	// UnpinVersions will remove all version pins for an environment.
	UnpinVersions(environmentName string) error
	// Delete will delete an environment from the state file
	Delete(environmentName string) error
	// initialize will overwite existing state with a new empty state file
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
		writer: newFileWriter(dir, "state.json"),
	}, nil
}

// package-private constructor, used in testing
func newWithBucket(_bucket bucket.Bucket, cfg statebucketConfig) *statebucket {
	return &statebucket{
		writer: newBucketWriter(_bucket, cfg),
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

		// make sure marshaled json includes an empty map so version pins is never nil when unmarshaled
		if environment.VersionPins == nil {
			environment.VersionPins = make(map[string]string)
		}
		_, exists := state.Environments[environment.Name]
		if exists {
			return StateFile{}, fmt.Errorf("can't add new environment %s, an environment by that name already exists", environment.Name)
		}
		state.Environments[environment.Name] = environment
		return state, nil
	})
}

func (s *statebucket) PinVersions(environmentName string, versionPins map[string]string) error {
	return s.writer.update(func(state StateFile) (StateFile, error) {
		environment, exists := state.Environments[environmentName]
		if !exists {
			return StateFile{}, fmt.Errorf("can't update environment %s, it does not exist in the state file", environmentName)
		}
		for service, version := range versionPins {
			environment.VersionPins[service] = version
		}
		state.Environments[environmentName] = environment
		return state, nil
	})
}

func (s *statebucket) UnpinVersions(environmentName string) error {
	return s.writer.update(func(state StateFile) (StateFile, error) {
		environment, exists := state.Environments[environmentName]
		if !exists {
			return StateFile{}, fmt.Errorf("can't update environment %s, it does not exist in the state file", environmentName)
		}
		environment.VersionPins = map[string]string{}
		state.Environments[environmentName] = environment
		return state, nil
	})
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

// populate a new empty statefile in the bucket
func (s *statebucket) initialize() error {
	return s.writer.write(StateFile{})
}
