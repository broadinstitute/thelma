package statebucket

import (
	"crypto/sha256"
	"fmt"
	"github.com/rs/zerolog/log"
	"k8s.io/utils/strings"
)

// backwardsCompatibleResourcePrefix generate a probably-unique 4-character resource prefix for the BEE of the form [a-z][a-z0-9]{3}.
func backwardsCompatibleResourcePrefix(environmentName string) string {
	return fmt.Sprintf("e%s", urpSuffix(environmentName))
}

// generateUniqueResourcePrefix generate a unique 4-character resource prefix for the BEE of the form [a-z][a-z0-9]{3}.
// We do this by:
// * taking first 3 chars of sha256sum of the environment name -- these become the last 3 chars of the URP
// * cycling through potential first letters from a-z until we find a URP that is not already in use
// (this algorithm is a dumb hack intended to last until we can do this better in Sherlock)
func generateUniqueResourcePrefix(environmentName string, stateFile StateFile) (string, error) {
	lastThreeChars := urpSuffix(environmentName)
	for i := 0; i < 26; i++ {
		firstLetter := 'a' + i

		// Don't generate resource prefixes that start with 'b' or 'e', which are used by older BEEs and fiabs.
		if firstLetter == 'b' || firstLetter == 'e' {
			continue
		}

		maybeURP := fmt.Sprintf("%c%s", firstLetter, lastThreeChars)

		if !isURPInUse(maybeURP, stateFile) {
			log.Debug().Msgf("Environment %s will use the following unique resource prefix: %q", environmentName, maybeURP)
			return maybeURP, nil
		}
	}

	return "", fmt.Errorf("exhausted all possible resource prefixes for environment name: %s (suffix: %s)", environmentName, lastThreeChars)
}

func isURPInUse(maybeURP string, stateFile StateFile) bool {
	for _, e := range stateFile.Environments {
		if e.UniqueResourcePrefix == maybeURP {
			log.Debug().Msgf("generated unique resource prefix %s is already in use by environment %s, will generate a new one", maybeURP, e.Name)
			return true
		}
	}
	return false
}

func urpSuffix(environmentName string) string {
	sha256sum := fmt.Sprintf("%x", sha256.Sum256([]byte(environmentName)))
	return strings.ShortenString(sha256sum, 3)
}
