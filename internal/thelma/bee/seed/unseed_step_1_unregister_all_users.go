package seed

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/terraapi"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func (s *seeder) unseedStep1UnregisterAllUsers(appReleases map[string]terra.AppRelease, opts UnseedOptions) error {
	log.Info().Msg("unregistering all users with Sam...")
	if sam, samPresent := appReleases["sam"]; samPresent {
		if !sam.Environment().Lifecycle().IsDynamic() {
			// Should never hit here because the caller checks for dynamic environment, but better safe than sorry
			panic("THIS SAM IS NOT IN A DYNAMIC ENVIRONMENT, REFUSING TO UNREGISTER ALL USERS")
		}

		k8s, err := s.clientFactory.Kubernetes().ForRelease(sam)
		if err != nil {
			return errors.Errorf("error getting Kubernetes client: %v", err)
		}

		config, err := s.configWithBasicDefaults()
		if err != nil {
			return errors.Errorf("error getting Sam's database info: %v", err)
		}

		secretName := config.Sam.Database.Credentials.KubernetesSecretName
		secret, err := k8s.CoreV1().Secrets(sam.Namespace()).Get(context.Background(), secretName, v1.GetOptions{})
		if err != nil {
			return errors.Errorf("error getting Sam's database credentials from Kubernetes secret %s/%s: %v", sam.Namespace(), secretName, err)
		}
		dbUsername, exists := secret.Data[config.Sam.Database.Credentials.KubernetesUsernameKey]
		if !exists {
			return errors.Errorf("secret at %s/%s didn't contain a %s key for Sam's database username", sam.Namespace(), secretName, config.Sam.Database.Credentials.KubernetesUsernameKey)
		}
		dbPassword, exists := secret.Data[config.Sam.Database.Credentials.KubernetesPasswordKey]
		if !exists {
			return errors.Errorf("secret at %s/%s didn't contain a %s key for Sam's database password", sam.Namespace(), secretName, config.Sam.Database.Credentials.KubernetesPasswordKey)
		}

		localPort, stopFunc, err := s.kubectl.PortForward(sam, fmt.Sprintf("service/%s", config.Sam.Database.Service), config.Sam.Database.Port)
		if err != nil {
			return errors.Errorf("error port-forwarding to Sam's database: %v", err)
		}
		defer func() { _ = stopFunc() }()

		db, err := sql.Open("pgx", fmt.Sprintf("user=%s password=%s host=localhost port=%d dbname=%s sslmode=disable",
			dbUsername, dbPassword, localPort, config.Sam.Database.Name))
		if err != nil {
			return errors.Errorf("error connecting to Sam's database: %v", err)
		}
		defer func() { _ = db.Close() }()

		err = db.Ping()
		if err != nil {
			return errors.Errorf("error pinging Sam's database: %v", err)
		}

		dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		rows, err := db.QueryContext(dbCtx, config.Sam.ListUserQuery)
		if err != nil {
			return errors.Errorf("error querying Sam's users: %v", err)
		}
		defer func() { _ = rows.Close() }()

		userEmailToID := make(map[string]string)
		for rows.Next() {
			var email, id string
			if err := rows.Scan(&email, &id); err != nil {
				if err = opts.handleErrorWithForce(errors.Errorf("error reading email/id for user: %v", err)); err != nil {
					return err
				}
			}
			userEmailToID[email] = id
		}
		if err := opts.handleErrorWithForce(rows.Err()); err != nil {
			return err
		}

		log.Info().Msgf("found %d users to remove", len(userEmailToID))

		if len(userEmailToID) > 0 {
			googleClient, err := s.googleAuthAs(sam)
			if err != nil {
				return errors.Errorf("couldn't prepare Google authentication as Sam's SA: %v", err)
			}
			terraClient, err := googleClient.Terra()
			if err != nil {
				return errors.Errorf("error authenticating to Google: %v", err)
			}

			samEmail := terraClient.GoogleUserinfo().Email
			if _, exists := userEmailToID[samEmail]; !exists {
				if err := opts.handleErrorWithForce(errors.Errorf("%s (Sam's SA) is not a Sam user and cannot unregister users", samEmail)); err != nil {
					return err
				}
			}

			var jobs []pool.Job
			for unsafeEmail, unsafeID := range userEmailToID {
				email := unsafeEmail
				id := unsafeID
				if email != samEmail {
					jobs = append(jobs, pool.Job{
						Name: email,
						Run: func(reporter pool.StatusReporter) error {
							var err error
							reporter.Update(pool.Status{
								Message: "Authenticating",
							})
							var googleClient google.Clients
							googleClient, err = s.googleAuthAs(sam)
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
								Message: "Unregistering",
							})
							_, _, err = terraClient.Sam(sam).UnregisterUser(id)
							if err := opts.handleErrorWithForce(err); err != nil {
								return errors.Errorf("error unregistering %s (%s): %v", email, id, err)
							}
							reporter.Update(pool.Status{
								Message: "Unregistered",
							})
							return nil
						},
					})
				} else {
					log.Debug().Msgf("skipping %s for now since it is %s's own user, can't delete it yet", id, samEmail)
				}
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

			if samId, exists := userEmailToID[samEmail]; exists {
				log.Info().Msgf("unregistering Sam's own %s user", samEmail)
				_, _, err = terraClient.Sam(sam).UnregisterUser(samId)
				if err := opts.handleErrorWithForce(err); err != nil {
					return errors.Errorf("error unregistering %s (%s): %v", samEmail, samId, err)
				}
			}
		}

	} else {
		log.Info().Msg("Sam not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}
