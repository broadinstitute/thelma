package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
)

func (cmd *seedCommand) step4RegisterTestUsers(thelma app.ThelmaApp, appReleases map[string]terra.AppRelease) error {
	log.Info().Msg("registering test users with Orch and Sam...")
	seedConfig, err := seed.ConfigWithTestUsers(thelma)
	if err != nil {
		return err
	}
	if orch, orchPresent := appReleases["firecloudorch"]; orchPresent {
		if sam, samPresent := appReleases["sam"]; samPresent {
			var users []seed.TestUser
			if sam.Cluster().ProjectSuffix() == "dev" {
				users = seedConfig.TestUsers.Dev
			} else if sam.Cluster().ProjectSuffix() == "qa" {
				users = seedConfig.TestUsers.QA
			} else {
				return fmt.Errorf("suffix %s of project %s maps to no test users", sam.Cluster().ProjectSuffix(), sam.Cluster().Project())
			}

			googleClient, err := seed.GoogleAuthAs(thelma, orch)
			if err != nil {
				return err
			}

			for index, user := range users {
				log.Info().Msgf("User %d - %s - registering", index+1, user.Email)
				terraClient, err := googleClient.SetSubject(user.Email).Terra()
				if err != nil {
					if cmd.handleErrorWithForce(err) != nil {
						return err
					}
					continue
				}
				_, _, err = terraClient.FirecloudOrch(orch).RegisterProfile(
					user.FirstName, user.LastName, user.Role, user.Email,
					"Hogwarts", "dsde",
					"Cambridge", "MA", "USA",
					"Remus Lupin", "Non-Profit")
				if err = cmd.handleErrorWithForce(err); err != nil {
					return err
				}
				log.Info().Msgf("User %d - %s - approving Terms of Service", index+1, user.Email)
				_, _, err = terraClient.Sam(sam).AcceptToS()
				if err = cmd.handleErrorWithForce(err); err != nil {
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
