package seed

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
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

		vault, err := s.clientFactory.Vault()
		if err != nil {
			return fmt.Errorf("error getting vault client: %v", err)
		}

		config, err := s.configWithBasicDefaults()
		if err != nil {
			return fmt.Errorf("error getting Sam's database info: %v", err)
		}

		secretPath := fmt.Sprintf(config.Sam.Database.Credentials.VaultPath, sam.Cluster().ProjectSuffix())
		secret, err := vault.Logical().Read(secretPath)
		if err != nil {
			return fmt.Errorf("error getting Sam's database credentials from %s: %v", secretPath, err)
		}
		dbUsername, exists := secret.Data[config.Sam.Database.Credentials.VaultUsernameKey]
		if !exists {
			return fmt.Errorf("secret at %s didn't contain a %s key for Sam's database username", secretPath, config.Sam.Database.Credentials.VaultUsernameKey)
		}
		dbPassword, exists := secret.Data[config.Sam.Database.Credentials.VaultPasswordKey]
		if !exists {
			return fmt.Errorf("secret at %s didn't contain a %s key for Sam's database password", secretPath, config.Sam.Database.Credentials.VaultPasswordKey)
		}

		localPort, stopFunc, err := s.kubectl.PortForward(sam, fmt.Sprintf("service/%s", config.Sam.Database.Service), config.Sam.Database.Port)
		if err != nil {
			return fmt.Errorf("error port-forwarding to Sam's database: %v", err)
		}
		defer func() { _ = stopFunc() }()

		db, err := sql.Open("pgx", fmt.Sprintf("user=%s password=%s host=localhost port=%d dbname=%s sslmode=disable",
			dbUsername, dbPassword, localPort, config.Sam.Database.Name))
		if err != nil {
			return fmt.Errorf("error connecting to Sam's database: %v", err)
		}
		defer func() { _ = db.Close() }()

		err = db.Ping()
		if err != nil {
			return fmt.Errorf("error pinging Sam's database: %v", err)
		}

		dbCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		rows, err := db.QueryContext(dbCtx, config.Sam.ListUserQuery)
		if err != nil {
			return fmt.Errorf("error querying Sam's users: %v", err)
		}
		defer func() { _ = rows.Close() }()

		userEmailToID := make(map[string]string)
		for rows.Next() {
			var email, id string
			if err := rows.Scan(&email, &id); err != nil {
				if err = opts.handleErrorWithForce(fmt.Errorf("error reading email/id for user: %v", err)); err != nil {
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
				return fmt.Errorf("couldn't prepare Google authentication as Sam's SA: %v", err)
			}
			terraClient, err := googleClient.Terra()
			if err != nil {
				return fmt.Errorf("error authenticating to Google: %v", err)
			}

			samEmail := terraClient.GoogleUserInfo().Email
			if _, exists := userEmailToID[samEmail]; !exists {
				if err := opts.handleErrorWithForce(fmt.Errorf("%s (Sam's SA) is not a Sam user and cannot unregister users", samEmail)); err != nil {
					return err
				}
			}

			for email, id := range userEmailToID {
				if email != samEmail {
					log.Info().Msgf("unregistering %s", email)
					_, _, err = terraClient.Sam(sam).UnregisterUser(id)
					if err := opts.handleErrorWithForce(err); err != nil {
						return fmt.Errorf("error unregistering %s (%s): %v", email, id, err)
					}
				} else {
					log.Debug().Msgf("skipping %s for now since it is %s's own user, can't delete it yet", id, samEmail)
				}
			}

			if samId, exists := userEmailToID[samEmail]; exists {
				log.Info().Msgf("unregistering Sam's own %s user", samEmail)
				_, _, err = terraClient.Sam(sam).UnregisterUser(samId)
				if err := opts.handleErrorWithForce(err); err != nil {
					return fmt.Errorf("error unregistering %s (%s): %v", samEmail, samId, err)
				}
			}
		}

	} else {
		log.Info().Msg("Sam not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}
