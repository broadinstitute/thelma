package seed

import (
	"bytes"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"strings"
)

func (s *seeder) seedStep6ExtraUser(appReleases map[string]terra.AppRelease, opts SeedOptions) error {
	log.Info().Msg("registering extra users with Orch and Sam...")
	if orch, orchPresent := appReleases["firecloudorch"]; orchPresent {
		if sam, samPresent := appReleases["sam"]; samPresent {
			for index, extraUser := range opts.Step6ExtraUser {
				log.Info().Msgf("Extra user %d - %s - registering", index+1, extraUser)

				if strings.ToLower(extraUser) == "set-adc" {
					log.Info().Msg("Running `gcloud auth application-default login`, opening your browser...")
					scopes := []string{
						"https://www.googleapis.com/auth/accounts.reauth",
						"https://www.googleapis.com/auth/cloud-platform",
						"https://www.googleapis.com/auth/sqlservice.login",
						"https://www.googleapis.com/auth/userinfo.email",
						"https://www.googleapis.com/auth/userinfo.profile",
						"openid",
					}
					err := s.shellRunner.Run(shell.Command{
						Prog: "gcloud",
						Args: []string{
							"auth",
							"application-default",
							"login",
							fmt.Sprintf("--scopes=%s", strings.Join(scopes, ",")),
						}})
					if err != nil {
						if err = opts.handleErrorWithForce(err); err != nil {
							return err
						}
						continue
					}
					log.Info().Msg("...done")
				}

				var googleClient google.Clients
				if strings.ToLower(extraUser) == "use-adc" || strings.ToLower(extraUser) == "set-adc" {
					googleClient = s.clientFactory.GoogleUsingADC(true)
				} else {
					g, err := s.googleAuthAs(orch)
					if err != nil {
						if err = opts.handleErrorWithForce(err); err != nil {
							return err
						}
						continue
					}
					googleClient = g
					googleClient.SetSubject(extraUser)
				}

				terraClient, err := googleClient.Terra()
				if err != nil {
					if err = opts.handleErrorWithForce(err); err != nil {
						return err
					}
					continue
				}

				firstName := terraClient.GoogleUserInfo().GivenName
				lastName := terraClient.GoogleUserInfo().FamilyName
				if firstName == "" || lastName == "" {
					log.Warn().Msg("Name missing from GoogleUserInfo with current credentials (probably ADC with default scopes)")
					var buffer bytes.Buffer
					err = s.shellRunner.Run(
						shell.Command{
							Prog: "bash",
							Args: []string{"-c", "id -P $(stat -f%Su /dev/console) | awk -F '[:]' '{print $8}'"}},
						func(opts *shell.RunOptions) {
							opts.Stdout = &buffer
						})
					if err == nil {
						name := strings.TrimSpace(buffer.String())
						log.Info().Msgf("Found \"%s\" on the local machine, substituting", name)
						nameParts := strings.Split(name, " ")
						if firstName == "" {
							firstName = nameParts[0]
						}
						if lastName == "" && len(nameParts) > 1 {
							lastName = nameParts[1]
						}
					}
					emailHandle := strings.Split(terraClient.GoogleUserInfo().Email, "@")[0]
					if firstName == "" {
						log.Info().Msgf("Using %s as the first name", emailHandle)
						firstName = emailHandle
					}
					if lastName == "" {
						log.Info().Msgf("Using %s as the last name", emailHandle)
						lastName = emailHandle
					}
				}

				_, _, err = terraClient.FirecloudOrch(orch).RegisterProfile(
					firstName, lastName,
					"Owner", terraClient.GoogleUserInfo().Email,
					"None", "None",
					"None", "None", "None",
					"None", "None")
				if err = opts.handleErrorWithForce(err); err != nil {
					return err
				}
				log.Info().Msgf("Extra user %d - %s - approving Terms of Service", index+1, extraUser)
				_, _, err = terraClient.Sam(sam).AcceptToS()
				if err = opts.handleErrorWithForce(err); err != nil {
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
