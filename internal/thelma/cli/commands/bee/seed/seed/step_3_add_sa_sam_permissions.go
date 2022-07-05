package seed

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
)

func (cmd *seedCommand) step3AddSaSamPermissions(thelma app.ThelmaApp, appReleases map[string]terra.AppRelease) error {
	log.Info().Msg("adding SAs of other apps to Sam permissions...")
	if sam, samPresent := appReleases["sam"]; samPresent {

		var emails []string

		if rawls, rawlsPresent := appReleases["rawls"]; rawlsPresent {
			log.Info().Msgf("will add Rawls SA permissions to %s", sam.Host())
			googleClient, err := seed.GoogleAuthAs(thelma, rawls)
			if err := cmd.handleErrorWithForce(err); err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err := cmd.handleErrorWithForce(err); err != nil {
				return err
			}
			emails = append(emails, terraClient.GoogleUserInfo().Email)
		} else {
			log.Info().Msg("Rawls not present in environment, skipping")
		}

		if leo, leoPresent := appReleases["leonardo"]; leoPresent {
			log.Info().Msgf("will add Leo SA permissions to %s", sam.Host())
			googleClient, err := seed.GoogleAuthAs(thelma, leo)
			if err := cmd.handleErrorWithForce(err); err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err := cmd.handleErrorWithForce(err); err != nil {
				return err
			}
			emails = append(emails, terraClient.GoogleUserInfo().Email)
		} else {
			log.Info().Msg("Leo not present in environment, skipping")
		}

		if importService, importServicePresent := appReleases["importservice"]; importServicePresent {
			log.Info().Msgf("will add Import Service SA permissions to %s", sam.Host())
			googleClient, err := seed.GoogleAuthAs(thelma, importService)
			if err := cmd.handleErrorWithForce(err); err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err := cmd.handleErrorWithForce(err); err != nil {
				return err
			}
			emails = append(emails, terraClient.GoogleUserInfo().Email)
		} else {
			log.Info().Msg("Import Service not present in environment, skipping")
		}

		googleClient, err := seed.GoogleAuthAs(thelma, sam)
		if err != nil {
			return err
		}
		terraClient, err := googleClient.Terra()
		if err := cmd.handleErrorWithForce(err); err != nil {
			return err
		}
		emails = append(emails, terraClient.GoogleUserInfo().Email)
		_, _, err = terraClient.Sam(sam).FcServiceAccounts(emails)
		if err != nil {
			return err
		}

	} else {
		log.Info().Msg("Sam not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}
