package seed

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/terraapi"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
)

func (s *seeder) seedStep3AddSaSamPermissions(appReleases map[string]terra.AppRelease, opts SeedOptions) error {
	log.Info().Msg("adding SAs of other apps to Sam permissions...")
	if sam, samPresent := appReleases["sam"]; samPresent {

		var emails []string
		if rawls, rawlsPresent := appReleases["rawls"]; rawlsPresent {
			log.Info().Msgf("will add Rawls SA permissions to %s", sam.Host())
			googleClient, err := s.googleAuthAs(rawls)
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			emails = append(emails, terraClient.GoogleUserInfo().Email)
		} else {
			log.Info().Msg("Rawls not present in environment, skipping")
		}

		if leo, leoPresent := appReleases["leonardo"]; leoPresent {
			log.Info().Msgf("will add Leo SA permissions to %s", sam.Host())
			googleClient, err := s.googleAuthAs(leo)
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			emails = append(emails, terraClient.GoogleUserInfo().Email)
		} else {
			log.Info().Msg("Leo not present in environment, skipping")
		}

		if importService, importServicePresent := appReleases["importservice"]; importServicePresent {
			log.Info().Msgf("will add Import Service SA permissions to %s", sam.Host())
			googleClient, err := s.googleAuthAs(importService)
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			emails = append(emails, terraClient.GoogleUserInfo().Email)
		} else {
			log.Info().Msg("Import Service not present in environment, skipping")
		}

		if workspaceManager, workspaceManagerPresent := appReleases["workspacemanager"]; workspaceManagerPresent {
			log.Info().Msgf("will add Workspace Manager SA permissions to %s", sam.Host())
			googleClient, err := s.googleAuthAs(workspaceManager)
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err := opts.handleErrorWithForce(err); err != nil {
				return err
			}
			emails = append(emails, terraClient.GoogleUserInfo().Email)
		} else {
			log.Info().Msg("Workspace Manager not present in environment, skipping")
		}

		googleClient, err := s.googleAuthAs(sam)
		if err != nil {
			return err
		}
		terraClient, err := googleClient.Terra()
		if err := opts.handleErrorWithForce(err); err != nil {
			return err
		}
		emails = append(emails, terraClient.GoogleUserInfo().Email)
		_, _, err = terraClient.Sam(sam).FcServiceAccounts(emails, "google", terraapi.GetPetPrivateKeyAction)
		if err != nil {
			return err
		}
		log.Info().Msg("creating azure cloud extension sam resource")
		_, _, err = terraClient.Sam(sam).CreateCloudExtension("azure")
		if err != nil {
			return err
		}

		_, _, err = terraClient.Sam(sam).FcServiceAccounts(emails, "azure", terraapi.GetPetManagedIdentityAction)
		if err != nil {
			return err
		}

	} else {
		log.Info().Msg("Sam not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}
