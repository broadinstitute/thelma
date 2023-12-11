package seed

import (
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/pkg/errors"
	"regexp"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
)

func (s *seeder) seedStep2RegisterSaProfiles(appReleases map[string]terra.AppRelease, opts SeedOptions) error {
	log.Info().Msg("registering app SA profiles with Orch...")
	if orch, orchPresent := appReleases["firecloudorch"]; orchPresent {
		var jobs []pool.Job

		jobs = append(jobs, pool.Job{
			Name: "firecloudorch SA",
			Run: func(reporter pool.StatusReporter) error {
				reporter.Update(pool.Status{Message: "Registering"})
				if err := opts.handleErrorWithForce(s._registerSaProfile(orch, orch)); err != nil {
					return err
				}
				reporter.Update(pool.Status{Message: "Registered"})
				return nil
			},
		})

		if rawls, rawlsPresent := appReleases["rawls"]; rawlsPresent {
			jobs = append(jobs, pool.Job{
				Name: "rawls SA",
				Run: func(reporter pool.StatusReporter) error {
					reporter.Update(pool.Status{Message: "Registering"})
					if err := opts.handleErrorWithForce(s._registerSaProfile(rawls, orch)); err != nil {
						return err
					}
					reporter.Update(pool.Status{Message: "Registered"})
					return nil
				},
			})
		} else {
			log.Info().Msg("Rawls not present in environment, skipping")
		}

		if sam, samPresent := appReleases["sam"]; samPresent {
			jobs = append(jobs, pool.Job{
				Name: "sam SA",
				Run: func(reporter pool.StatusReporter) error {
					reporter.Update(pool.Status{Message: "Registering"})
					if err := opts.handleErrorWithForce(s._registerSaProfile(sam, orch)); err != nil {
						return err
					}
					reporter.Update(pool.Status{Message: "Registered"})
					return nil
				},
			})
		} else {
			log.Info().Msg("Sam not present in environment, skipping")
		}

		if leo, leoPresent := appReleases["leonardo"]; leoPresent {
			jobs = append(jobs, pool.Job{
				Name: "leonardo SA",
				Run: func(reporter pool.StatusReporter) error {
					reporter.Update(pool.Status{Message: "Registering"})
					if err := opts.handleErrorWithForce(s._registerSaProfile(leo, orch)); err != nil {
						return err
					}
					reporter.Update(pool.Status{Message: "Registered"})
					return nil
				},
			})
		} else {
			log.Info().Msg("Leo not present in environment, skipping")
		}

		if importService, importServicePresent := appReleases["importservice"]; importServicePresent {
			jobs = append(jobs, pool.Job{
				Name: "importservice SA",
				Run: func(reporter pool.StatusReporter) error {
					reporter.Update(pool.Status{Message: "Registering"})
					if err := opts.handleErrorWithForce(s._registerSaProfile(importService, orch)); err != nil {
						return err
					}
					reporter.Update(pool.Status{Message: "Registered"})
					return nil
				},
			})
		} else {
			log.Info().Msg("Import Service not present in environment, skipping")
		}

		if workspaceManager, workspaceManagerPresent := appReleases["workspacemanager"]; workspaceManagerPresent {
			jobs = append(jobs, pool.Job{
				Name: "workspacemanager SA",
				Run: func(reporter pool.StatusReporter) error {
					reporter.Update(pool.Status{Message: "Registering"})
					if err := opts.handleErrorWithForce(s._registerSaProfile(workspaceManager, orch)); err != nil {
						return err
					}
					reporter.Update(pool.Status{Message: "Registered"})
					return nil
				},
			})
		} else {
			log.Info().Msg("Workspace Manager not present in environment, skipping")
		}

		if tsps, tspsPresent := appReleases["tsps"]; tspsPresent {
			jobs = append(jobs, pool.Job{
				Name: "tsps SA",
				Run: func(reporter pool.StatusReporter) error {
					reporter.Update(pool.Status{Message: "Registering"})
					if err := opts.handleErrorWithForce(s._registerSaProfile(tsps, orch)); err != nil {
						return err
					}
					reporter.Update(pool.Status{Message: "Registered"})
					return nil
				},
			})
		} else {
			log.Info().Msg("TSPS not present in environment, skipping")
		}

		err := pool.New(jobs, func(o *pool.Options) {
			o.NumWorkers = opts.RegistrationParallelism
			o.Summarizer.Enabled = true
			o.Metrics.Enabled = false
			o.StopProcessingOnError = !opts.Force
		}).Execute()
		if err = opts.handleErrorWithForce(err); err != nil {
			return err
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
	_, _, err = terraClient.FirecloudOrch(orch).RegisterProfile("None", "None", "None", terraClient.GoogleUserinfo().Email, "None", "None", "None", "None", "None", "None", "None")

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
