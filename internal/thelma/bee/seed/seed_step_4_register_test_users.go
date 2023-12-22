package seed

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/terraapi"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
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

			var jobs []pool.Job
			for _, unsafe := range users {
				user := unsafe
				jobs = append(jobs, pool.Job{
					Name: user.Email,
					Run: func(reporter pool.StatusReporter) error {
						var err error
						reporter.Update(pool.Status{
							Message: "Authenticating",
						})
						var googleClient google.Clients
						googleClient, err = s.googleAuthAs(orch, func(options *google.Options) {
							options.Subject = user.Email
						})
						if err = opts.handleErrorWithForce(err); err != nil {
							return err
						}
						var terraClient terraapi.TerraClient
						terraClient, err = googleClient.Terra()
						if err = opts.handleErrorWithForce(err); err != nil {
							return err
						}
						terraClient.SetPoolStatusReporter(reporter)
						reporter.Update(pool.Status{
							Message: "Registering",
						})
						_, _, err = terraClient.FirecloudOrch(orch).RegisterProfile(
							user.FirstName, user.LastName, user.Role, user.Email,
							"Hogwarts", "dsde",
							"Cambridge", "MA", "USA",
							"Remus Lupin", "Non-Profit")
						if err = opts.handleErrorWithForce(err); err != nil {
							return err
						}
						reporter.Update(pool.Status{
							Message: "Approving TOS",
						})
						_, _, err = terraClient.Sam(sam).AcceptToS()
						if err = opts.handleErrorWithForce(err); err != nil {
							return err
						}
						reporter.Update(pool.Status{
							Message: "Registered",
						})
						return nil
					},
				})
			}

			err = pool.New(jobs, func(o *pool.Options) {
				o.NumWorkers = opts.RegistrationParallelism
				o.LogSummarizer.Enabled = true
				o.Metrics.Enabled = false
				o.StopProcessingOnError = !opts.Force
			}).Execute()
			if err = opts.handleErrorWithForce(err); err != nil {
				return err
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
