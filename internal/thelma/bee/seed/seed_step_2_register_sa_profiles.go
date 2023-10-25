package seed

import (
	"github.com/pkg/errors"
	"regexp"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
)

func (s *seeder) seedStep2RegisterSaProfiles(appReleases map[string]terra.AppRelease, opts SeedOptions) error {
	log.Info().Msg("registering app SA profiles with Orch...")
	if orch, orchPresent := appReleases["firecloudorch"]; orchPresent {

		log.Info().Msgf("registering Orch SA profile with %s", orch.Host())
		err := opts.handleErrorWithForce(s._registerSaProfile(orch, orch))
		if err != nil {
			return err
		}

		if rawls, rawlsPresent := appReleases["rawls"]; rawlsPresent {
			log.Info().Msgf("registering Rawls SA profile with %s", orch.Host())
			err = opts.handleErrorWithForce(s._registerSaProfile(rawls, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Rawls not present in environment, skipping")
		}

		if sam, samPresent := appReleases["sam"]; samPresent {
			log.Info().Msgf("registering Sam SA profile with %s", orch.Host())
			err = opts.handleErrorWithForce(s._registerSaProfile(sam, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Sam not present in environment, skipping")
		}

		if leo, leoPresent := appReleases["leonardo"]; leoPresent {
			log.Info().Msgf("registering Leo SA profile with %s", orch.Host())
			err = opts.handleErrorWithForce(s._registerSaProfile(leo, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Leo not present in environment, skipping")
		}

		if importService, importServicePresent := appReleases["importservice"]; importServicePresent {
			log.Info().Msgf("registering Import Service SA profile with %s", orch.Host())
			err = opts.handleErrorWithForce(s._registerSaProfile(importService, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Import Service not present in environment, skipping")
		}

		if workspaceManager, workspaceManagerPresent := appReleases["workspacemanager"]; workspaceManagerPresent {
			log.Info().Msgf("registering Workspace Manager SA profile with %s", orch.Host())
			err = opts.handleErrorWithForce(s._registerSaProfile(workspaceManager, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("Workspace Manager not present in environment, skipping")
		}

		if tsps, tspsPresent := appReleases["tsps"]; tspsPresent {
			log.Info().Msgf("registering TSPS SA profile with %s", orch.Host())
			err = opts.handleErrorWithForce(s._registerSaProfile(tsps, orch))
			if err != nil {
				return err
			}
		} else {
			log.Info().Msg("TSPS not present in environment, skipping")
		}

	} else {
		log.Info().Msg("Orch not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}

func (s *seeder) _registerSaProfile(appRelease terra.AppRelease, orch terra.AppRelease) error {
	googleClient, err := s.googleAuthAs(appRelease)
	if err != nil {
		return err
	}
	terraClient, err := googleClient.Terra()
	if err != nil {
		return err
	}
	_, _, err = terraClient.FirecloudOrch(orch).RegisterProfile("None", "None", "None", terraClient.GoogleUserInfo().Email, "None", "None", "None", "None", "None", "None", "None")

	return _ignore409Conflict(err)
}

func _ignore409Conflict(maybe409Err error) error {
	if maybe409Err == nil {
		return nil
	}

	pattern := "(?s)409 [Cc]onflict.*[Uu]ser.*already (?:exists|registered)"
	matches, err := regexp.MatchString(pattern, maybe409Err.Error())

	if err != nil {
		panic(errors.Errorf("invalid regular expression %q: %v", pattern, err))
	}

	if !matches {
		return maybe409Err
	}

	log.Warn().Err(maybe409Err).Msgf("409 conflict encountered while registering user; ignoring")
	return nil
}
