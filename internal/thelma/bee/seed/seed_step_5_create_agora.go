package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
)

func (s *seeder) seedStep5CreateAgora(appReleases map[string]terra.AppRelease, opts SeedOptions) error {
	log.Info().Msg("creating Agora methods repository with Orch...")
	seedConfig, err := s.configWithAgoraData()
	if err != nil {
		return err
	}
	if orch, orchPresent := appReleases["firecloudorch"]; orchPresent {
		if _, agoraPresent := appReleases["agora"]; agoraPresent {
			var acls []AgoraPermission
			if orch.Cluster().ProjectSuffix() == "dev" {
				acls = seedConfig.Agora.Permissions.Dev
			} else if orch.Cluster().ProjectSuffix() == "qa" {
				acls = seedConfig.Agora.Permissions.QA
			} else {
				err = fmt.Errorf("suffix %s of project %s maps to not Agora ACLs", orch.Cluster().ProjectSuffix(), orch.Cluster().Project())
				if err = opts.handleErrorWithForce(err); err != nil {
					return err
				}
			}

			googleClient, err := s.googleAuthAs(orch)
			if err != nil {
				return err
			}
			terraClient, err := googleClient.Terra()
			if err != nil {
				return err
			}
			orchClient := terraClient.FirecloudOrch(orch)

			// Go doesn't have a set primitive; struct{}{} doesn't take up any memory space, so it's close enough
			observedNamespaces := make(map[string]struct{})

			for index, method := range seedConfig.Agora.Methods {
				log.Info().Msgf("Method %d - %s - adding", index+1, method.Name)
				observedNamespaces[method.Namespace] = struct{}{}
				_, _, err = orchClient.AgoraMakeMethod(method)
				if err != nil {
					if err = opts.handleErrorWithForce(err); err != nil {
						return err
					}
					continue
				}
				if len(acls) > 0 {
					log.Info().Msgf("Method %d - %s - setting ACLs", index+1, method.Name)
					_, _, err = orchClient.AgoraSetMethodACLs(method.Name, method.Namespace, acls)
					if err = opts.handleErrorWithForce(err); err != nil {
						return err
					}
				} else {
					log.Info().Msg("No ACLs to set for methods, skipping")
				}
			}

			for index, config := range seedConfig.Agora.Configurations {
				log.Info().Msgf("Configuration %d - %s - adding", index+1, config.Name)
				observedNamespaces[config.Namespace] = struct{}{}
				_, _, err = orchClient.AgoraMakeConfig(config)
				if err != nil {
					if err = opts.handleErrorWithForce(err); err != nil {
						return err
					}
					continue
				}
				if len(acls) > 0 {
					log.Info().Msgf("Configuration %d - %s - setting ACLs", index+1, config.Name)
					_, _, err = orchClient.AgoraSetConfigACLs(config.Name, config.Namespace, acls)
					if err = opts.handleErrorWithForce(err); err != nil {
						return err
					}
				} else {
					log.Info().Msg("No ACLs to set for configurations, skipping")
				}
			}

			if len(acls) > 0 {
				for namespace := range observedNamespaces {
					log.Info().Msgf("Namespace %s - setting ACLs", namespace)
					_, _, err = orchClient.AgoraSetNamespaceACLs(namespace, acls)
					if err = opts.handleErrorWithForce(err); err != nil {
						return err
					}
				}
			} else {
				log.Info().Msg("No ACLs to set for namespaces, skipping")
			}

		} else {
			log.Info().Msg("Agora not present in environment, skipping all")
		}
	} else {
		log.Info().Msg("Orch not present in environment, skipping all")
	}
	log.Info().Msg("...done")
	return nil
}
