package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"regexp"
)

func (cmd *seedCommand) step2RegisterSaProfiles(thelma app.ThelmaApp, appReleases map[string]terra.AppRelease) error {
	log.Info().Msg("registering app SA profiles with Orch...")
	if orch, orchPresent := appReleases["firecloudorch"]; orchPresent {

		log.Info().Msgf("registering Orch SA profile with %s", orch.Host())
		err := cmd.handleErrorWithForce(_registerSaProfile(thelma, orch, orch))
		if err != nil {
			return err
		}

		if rawls, rawlsPresent := appReleases["rawls"]; rawlsPresent {
			log.Info().Msgf("registering Rawls SA profile with %s", orch.Host())
			err = cmd.handleErrorWithForce(_registerSaProfile(thelma, rawls, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Rawls not present in environment, skipping")
		}

		if sam, samPresent := appReleases["sam"]; samPresent {
			log.Info().Msgf("registering Sam SA profile with %s", orch.Host())
			err = cmd.handleErrorWithForce(_registerSaProfile(thelma, sam, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Sam not present in environment, skipping")
		}

		if leo, leoPresent := appReleases["leonardo"]; leoPresent {
			log.Info().Msgf("registering Leo SA profile with %s", orch.Host())
			err = cmd.handleErrorWithForce(_registerSaProfile(thelma, leo, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Leo not present in environment, skipping")
		}

		if importService, importServicePresent := appReleases["importservice"]; importServicePresent {
			log.Info().Msgf("registering Import Service SA profile with %s", orch.Host())
			err = cmd.handleErrorWithForce(_registerSaProfile(thelma, importService, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Import Service not present in environment, skipping")
		}

		if workspaceManager, workspaceManagerPresent := appReleases["workspacemanager"]; workspaceManagerPresent {
			log.Info().Msgf("registering Workspace Manager SA profile with %s", orch.Host())
			err = cmd.handleErrorWithForce(_registerSaProfile(thelma, workspaceManager, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Workspace Manager not present in environment, skipping")
		}

	} else {
		log.Info().Msg("Orch not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}

func _registerSaProfile(thelma app.ThelmaApp, appRelease terra.AppRelease, orch terra.AppRelease) error {
	googleClient, err := seed.GoogleAuthAs(thelma, appRelease)
	if err != nil {
		return err
	}
	terraClient, err := googleClient.Terra()
	if err != nil {
		return err
	}
	_, _, err = terraClient.FirecloudOrch(orch).RegisterProfile("None", "None", "None", terraClient.GoogleUserInfo().Email, "None", "None", "None", "None", "None", "None", "None")

	return ignore409Conflict(err)
}

func ignore409Conflict(maybe409Err error) error {
	if maybe409Err == nil {
		return nil
	}

	pattern := "(?s)409 [Cc]onflict.*[Uu]ser.*already exists"
	matches, err := regexp.MatchString(pattern, maybe409Err.Error())

	if err != nil {
		panic(fmt.Errorf("invalid regular expression %q: %v", pattern, maybe409Err))
	}

	if !matches {
		return maybe409Err
	}

	log.Warn().Err(maybe409Err).Msgf("409 conflict encountered while registering user; ignoring")
	return nil
}
