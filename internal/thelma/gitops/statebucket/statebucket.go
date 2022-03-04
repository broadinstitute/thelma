package statebucket

import (
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/gcp/bucket"
	"github.com/broadinstitute/thelma/internal/thelma/utils/gcp/bucket/lock"
	"github.com/rs/zerolog/log"
	"time"
)

const bucketName = "thelma-state"
const stateObject = "state.json"
const lockObject = ".update.lk"
const lockMaxWait = 30 * time.Second
const lockExpiresAfter = 300 * time.Second

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
}

// New returns a new statebucket
func New() (StateBucket, error) {
	_bucket, err := bucket.NewBucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("error initializing state bucket %s: %v", bucketName, err)
	}

	return newWithBucket(_bucket), nil
}

// package-private constructor, used in testing
func newWithBucket(_bucket bucket.Bucket) *statebucket {
	return &statebucket{
		objectName: stateObject,
		bucket:     _bucket,
	}
}

type statebucket struct {
	objectName string
	bucket     bucket.Bucket
}

func (s *statebucket) Environments() ([]DynamicEnvironment, error) {
	state, err := s.loadState()
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
	return result, nil
}

func (s *statebucket) Add(environment DynamicEnvironment) error {
	return s.transformState(func(state *StateFile) error {
		if state.Environments == nil {
			state.Environments = make(map[string]DynamicEnvironment)
		}

		// make sure marshaled json includes an empty map so version pins is never nil when unmarshaled
		if environment.VersionPins == nil {
			environment.VersionPins = make(map[string]string)
		}
		_, exists := state.Environments[environment.Name]
		if exists {
			return fmt.Errorf("can't add new environment %s, an environment by that name already exists", environment.Name)
		}
		state.Environments[environment.Name] = environment
		return nil
	})
}

func (s *statebucket) PinVersions(environmentName string, versionPins map[string]string) error {
	return s.transformState(func(state *StateFile) error {
		environment, exists := state.Environments[environmentName]
		if !exists {
			return fmt.Errorf("can't update environment %s, it does not exist in the state file", environmentName)
		}
		for service, version := range versionPins {
			environment.VersionPins[service] = version
		}
		state.Environments[environmentName] = environment
		return nil
	})
}

func (s *statebucket) UnpinVersions(environmentName string) error {
	return s.transformState(func(state *StateFile) error {
		environment, exists := state.Environments[environmentName]
		if !exists {
			return fmt.Errorf("can't update environment %s, it does not exist in the state file", environmentName)
		}
		environment.VersionPins = map[string]string{}
		state.Environments[environmentName] = environment
		return nil
	})
}

func (s *statebucket) Delete(environmentName string) error {
	return s.transformState(func(state *StateFile) error {
		_, exists := state.Environments[environmentName]
		if !exists {
			return fmt.Errorf("can't delete environment %s, it does not exist in the state file", environmentName)
		}
		delete(state.Environments, environmentName)
		return nil
	})
}

func (s *statebucket) initialize() error {
	content, err := json.Marshal(StateFile{})
	if err != nil {
		return fmt.Errorf("error marshalling empty state file: %v", err)
	}
	err = s.withLock(func() error {
		return s.bucket.Write(stateObject, content)
	})
	if err != nil {
		return fmt.Errorf("error initializing state file: %v", err)
	}
	return nil
}

func (s *statebucket) loadState() (StateFile, error) {
	var result StateFile
	data, err := s.bucket.Read(stateObject)

	if err != nil {
		return result, fmt.Errorf("error reading state file: %v", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("error unmarshalling state file: %v\nContent:\n%s", err, string(data))
	}

	return result, nil
}

func (s *statebucket) transformState(transformFn func(state *StateFile) error) error {
	err := s.withLock(func() error {
		return s.transformStateUnsafe(transformFn)
	})
	if err != nil {
		return fmt.Errorf("error updating state file: %v", err)
	}
	return nil
}

func (s *statebucket) transformStateUnsafe(transformFn func(state *StateFile) error) error {
	state, err := s.loadState()
	if err != nil {
		return err
	}

	if err := transformFn(&state); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("error marshalling state file: %v", err)
	}

	if err := s.bucket.Write(stateObject, data); err != nil {
		return fmt.Errorf("error writing state file: %v", err)
	}

	return nil
}

func (s *statebucket) withLock(fn func() error) error {
	locker := s.bucket.NewLocker(lockObject, lockMaxWait, func(options *lock.Options) {
		options.ExpiresAfter = lockExpiresAfter
	})

	lockId, err := locker.Lock()
	if err != nil {
		return err
	}

	fnErr := fn()

	err = locker.Unlock(lockId)
	if err != nil {
		log.Error().Err(err).Msgf("error releasing lock %s: %v", lockObject, err)
	}

	// if we got a callback error, return it, else return lock release error
	if fnErr != nil {
		return fnErr
	}
	return err
}
