package seed

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (s *seeder) seedStep4RegisterTestUsers(appReleases map[string]terra.AppRelease, opts SeedOptions) error {
	log.Info().Msg("registering test users with Orch and Sam...")
	seedConfig, err := s.configWithTestUsers()
	if err != nil {
		return err
	}
	if orch, orchPresent := appReleases["firecloudorch"]; orchPresent {
		if sam, samPresent := appReleases["sam"]; samPresent {
			var users []TestUser
			if sam.Cluster().ProjectSuffix() == "dev" {
				users = seedConfig.TestUsers.Dev
			} else if sam.Cluster().ProjectSuffix() == "qa" {
				users = seedConfig.TestUsers.QA
			} else {
				return errors.Errorf("suffix %s of project %s maps to no test users", sam.Cluster().ProjectSuffix(), sam.Cluster().Project())
			}

			googleClient, err := s.googleAuthAs(orch)
			if err != nil {
				return err
			}

			for index, user := range users {
				log.Info().Msgf("User %d - %s - registering", index+1, user.Email)
				terraClient, err := googleClient.SetSubject(user.Email).Terra()
				if err != nil {
					if opts.handleErrorWithForce(err) != nil {
						return err
					}
					continue
				}
				_, _, err = terraClient.FirecloudOrch(orch).RegisterProfile(
					user.FirstName, user.LastName, user.Role, user.Email,
					"Hogwarts", "dsde",
					"Cambridge", "MA", "USA",
					"Remus Lupin", "Non-Profit")
				if err = opts.handleErrorWithForce(err); err != nil {
					return err
				}
				log.Info().Msgf("User %d - %s - approving Terms of Service", index+1, user.Email)
				_, _, err = terraClient.Sam(sam).AcceptToS()
				if err = opts.handleErrorWithForce(err); err != nil {
					return err
				}
			}
		} else {
			log.Info().Msg("Sam not present in environment, skipping all")
		}
	} else {
		log.Info().Msg("Orch not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}
