package unseed

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func (cmd *unseedCommand) step1UnregisterAllUsers(thelma app.ThelmaApp, appReleases map[string]terra.AppRelease) error {
	log.Info().Msg("unregistering all users with Sam...")
	if sam, samPresent := appReleases["sam"]; samPresent {
		if !sam.Environment().Lifecycle().IsDynamic() {
			// Should never hit here because the caller checks for dynamic environment, but better safe than sorry
			panic("THIS SAM IS NOT IN A DYNAMIC ENVIRONMENT, REFUSING TO UNREGISTER ALL USERS")
		}

		kubectl, err := thelma.Clients().Google().Kubectl()
		if err != nil {
			return fmt.Errorf("error getting kubectl client: %v", err)
		}

		vault, err := thelma.Clients().Vault()
		if err != nil {
			return fmt.Errorf("error getting vault client: %v", err)
		}

		config, err := seed.ConfigWithBasicDefaults(thelma)
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

		localPort, stopFunc, err := kubectl.PortForward(sam, fmt.Sprintf("service/%s", config.Sam.Database.Service), config.Sam.Database.Port)
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
				if err = cmd.handleErrorWithForce(fmt.Errorf("error reading email/id for user: %v", err)); err != nil {
					return err
				}
			}
			userEmailToID[email] = id
		}
		if err := cmd.handleErrorWithForce(rows.Err()); err != nil {
			return err
		}

		log.Info().Msgf("found %d users to remove", len(userEmailToID))

		if len(userEmailToID) > 0 {
			googleClient, err := seed.GoogleAuthAs(thelma, sam)
			if err != nil {
				return fmt.Errorf("couldn't prepare Google authentication as Sam's SA: %v", err)
			}
			terraClient, err := googleClient.Terra()
			if err != nil {
				return fmt.Errorf("error authenticating to Google: %v", err)
			}

			samEmail := terraClient.GoogleUserInfo().Email
			if _, exists := userEmailToID[samEmail]; !exists {
				if err := cmd.handleErrorWithForce(fmt.Errorf("%s (Sam's SA) is not a Sam user and cannot unregister users", samEmail)); err != nil {
					return err
				}
			}

			for email, id := range userEmailToID {
				if email != samEmail {
					log.Info().Msgf("unregistering %s", email)
					_, _, err = terraClient.Sam(sam).UnregisterUser(id)
					if err := cmd.handleErrorWithForce(err); err != nil {
						return fmt.Errorf("error unregistering %s (%s): %v", email, id, err)
					}
				} else {
					log.Debug().Msgf("skipping %s for now since it is %s's own user, can't delete it yet", id, samEmail)
				}
			}

			if samId, exists := userEmailToID[samEmail]; exists {
				log.Info().Msgf("unregistering Sam's own %s user", samEmail)
				_, _, err = terraClient.Sam(sam).UnregisterUser(samId)
				if err := cmd.handleErrorWithForce(err); err != nil {
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
