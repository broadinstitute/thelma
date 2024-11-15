package seed

import (
	"context"
	"fmt"
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const configKey = "seed"

type TestUser struct {
	Role      string
	FirstName string
	LastName  string
	Email     string
}

type AgoraPayload struct {
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	Synopsis      string `json:"synopsis"`
	Documentation string `json:"documentation"`
	Payload       string `json:"payload"`
	EntityType    string `json:"entityType"`
}

type AgoraPermission struct {
	User string `json:"user"`
	Role string `json:"role"`
}

type seedConfig struct {
	Auth struct {
		Rawls struct {
			KubernetesSecretName string `default:"rawls-sa-secret"`
			KubernetesSecretKey  string `default:"rawls-account.json"`
		}
		Sam struct {
			KubernetesSecretName string `default:"sam-sa-secret"`
			KubernetesSecretKey  string `default:"sam-account.json"`
		}
		Leonardo struct {
			KubernetesSecretName string `default:"leonardo-sa-secret"`
			KubernetesSecretKey  string `default:"leonardo-account.json"`
		}
		FirecloudOrch struct {
			KubernetesSecretName string `default:"firecloudorch-sa-secret"`
			KubernetesSecretKey  string `default:"firecloud-account.json"`
		}
		WorkspaceManager struct {
			// WSM dev SA used for both Dev and QA BEEs, as of 7/13/2022
			KubernetesSecretName string `default:"workspacemanager-sa-secret"`
			KubernetesSecretKey  string `default:"service-account.json"`
		}
		TSPS struct {
			KubernetesSecretName string `default:"tsps-sa-secret"`
			KubernetesSecretKey  string `default:"service-account.json"`
		}
		Teaspoons struct {
			KubernetesSecretName string `default:"teaspoons-sa-secret"`
			KubernetesSecretKey  string `default:"service-account.json"`
		}
		Datarepo struct {
			KubernetesSecretName string `default:"jade-sa"`
			KubernetesSecretKey  string `default:"datareposerviceaccount"`
		}
	}
	TestUsers struct {
		Dev []TestUser
		QA  []TestUser
	}
	Agora struct {
		Methods        []AgoraPayload
		Configurations []AgoraPayload
		Permissions    struct {
			Dev []AgoraPermission
			QA  []AgoraPermission
		}
	}
	Elasticsearch struct {
		Service string `default:"elasticsearch-0"`
	}
	Sam struct {
		Database struct {
			Service     string `default:"sam-postgres-service"`
			Name        string `default:"sam"`
			Port        int    `default:"5432"`
			Credentials struct {
				KubernetesSecretName  string `default:"sam-db-creds-eso"`
				KubernetesUsernameKey string `default:"username"`
				KubernetesPasswordKey string `default:"password"`
			}
		}
		ListUserQuery string `default:"SELECT email, id FROM sam_user"`
	}
}

func (s *seeder) googleAuthAs(appRelease terra.AppRelease, options ...google.Option) (google.Clients, error) {
	config, err := s.configWithBasicDefaults()
	if err != nil {
		return nil, err
	}
	var secretName, secretKey string
	switch appRelease.Name() {
	case "rawls":
		secretName = config.Auth.Rawls.KubernetesSecretName
		secretKey = config.Auth.Rawls.KubernetesSecretKey
	case "sam":
		secretName = config.Auth.Sam.KubernetesSecretName
		secretKey = config.Auth.Sam.KubernetesSecretKey
	case "leonardo":
		secretName = config.Auth.Leonardo.KubernetesSecretName
		secretKey = config.Auth.Leonardo.KubernetesSecretKey
	case "firecloudorch":
		secretName = config.Auth.FirecloudOrch.KubernetesSecretName
		secretKey = config.Auth.FirecloudOrch.KubernetesSecretKey
	case "workspacemanager":
		secretName = config.Auth.WorkspaceManager.KubernetesSecretName
		secretKey = config.Auth.WorkspaceManager.KubernetesSecretKey
        case "teaspoons":
		secretName = config.Auth.Teaspoons.KubernetesSecretName
		secretKey = config.Auth.Teaspoons.KubernetesSecretKey		
	case "tsps":
		secretName = config.Auth.TSPS.KubernetesSecretName
		secretKey = config.Auth.TSPS.KubernetesSecretKey
	case "datarepo":
		secretName = config.Auth.Datarepo.KubernetesSecretName
		secretKey = config.Auth.Datarepo.KubernetesSecretKey
	default:
		return nil, errors.Errorf("thelma doesn't know how to authenticate as %s", appRelease.Name())
	}
	if strings.ContainsRune(secretName, '%') {
		secretName = fmt.Sprintf(secretName, appRelease.Cluster().ProjectSuffix())
	}

	k8s, err := s.clientFactory.Kubernetes().ForRelease(appRelease)
	if err != nil {
		return nil, errors.Errorf("failed to construct K8s client for release %s: %v", appRelease.FullName(), err)
	}
	secret, err := k8s.CoreV1().Secrets(appRelease.Namespace()).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Errorf("failed to access %s service account key secret %s: %v", appRelease.FullName(), secretName, err)
	}
	saKeyJSON, exists := secret.Data[secretKey]
	if !exists || len(saKeyJSON) == 0 {
		return nil, errors.Errorf("failed to %s service account key secret %s missing key %s: %v", appRelease.FullName(), secretName, secretKey, err)
	}

	log.Debug().Msgf("Successfully downloaded SA key for %s from secret %s (key %s)", appRelease.FullName(), secretName, secretKey)

	return s.clientFactory.Google(
		append(options, google.OptionForceSAKey(saKeyJSON))...,
	), nil
}

func (s *seeder) configWithBasicDefaults() (seedConfig, error) {
	var config seedConfig
	err := s.config.Unmarshal(configKey, &config)
	if err != nil {
		return config, errors.Errorf("error reading seed config: %v", err)
	}
	return config, nil
}

func (s *seeder) configWithTestUsers() (seedConfig, error) {
	config, err := s.configWithBasicDefaults()
	if err != nil {
		return config, err
	}
	if len(config.TestUsers.Dev) > 0 {
		log.Info().Msg("test users for dev already present in config, skipping defaults")
	} else {
		config.TestUsers.Dev = []TestUser{
			// "Initial User"
			{"Owner", "Hermione", "Granger", "hermione.owner@test.firecloud.org"},

			{"Professor", "Albus", "Dumbledore", "dumbledore.admin@test.firecloud.org"},
			{"Professor", "Lord", "Voldemort", "voldemort.admin@test.firecloud.org"},
			{"Professor", "Minerva", "McGonagall", "mcgonagall.curator@test.firecloud.org"},
			{"Professor", "Severus", "Snape", "snape.curator@test.firecloud.org"},
			{"Student", "Harry", "Potter", "harry.potter@test.firecloud.org"},
			{"Student", "Ron", "Weasley", "ron.weasley@test.firecloud.org"},
			{"Student", "Draco", "Malfoy", "draco.malfoy@test.firecloud.org"},
			{"Researcher", "Fred", "Weasley", "fred.authdomain@test.firecloud.org"},
			{"Researcher", "George", "Weasley", "george.authdomain@test.firecloud.org"},
			{"Researcher", "Bill", "Weasley", "bill.authdomain@test.firecloud.org"},
		}
	}
	if len(config.TestUsers.QA) > 0 {
		log.Info().Msg("test users for QA already present in config, skipping defaults")
	} else {
		config.TestUsers.QA = []TestUser{
			// "Initial User"
			{"Owner", "Hermione", "Granger", "hermione.owner@quality.firecloud.org"},

			// Admins
			{"Professor", "Albus", "Dumbledore", "dumbledore.admin@quality.firecloud.org"},
			{"Professor", "Lord", "Voldemort", "voldemort.admin@quality.firecloud.org"},

			// Curators
			{"Professor", "Minerva", "McGonagall", "mcgonagall.curator@quality.firecloud.org"},
			{"Professor", "Remus", "Lupin", "lupin.curator@quality.firecloud.org"},
			{"Professor", "Filius", "Flitwick", "flitwick.curator@quality.firecloud.org"},
			{"Professor", "Rubeus", "Hagrid", "hagrid.curator@quality.firecloud.org"},
			{"Professor", "Severus", "Snape", "snape.curator@quality.firecloud.org"},

			//Project Owners
			{"Owner", "Sirius", "Black", "sirius.owner@quality.firecloud.org"},
			{"Owner", "Nymphadora", "Tonks", "tonks.owner@quality.firecloud.org"},

			// Students
			{"Student", "Harry", "Potter", "harry.potter@quality.firecloud.org"},
			{"Student", "Ron", "Weasley", "ron.weasley@quality.firecloud.org"},
			{"Student", "Draco", "Malfoy", "draco.malfoy@quality.firecloud.org"},
			{"Student", "Lavender", "Brown", "lavender.brown@quality.firecloud.org"},
			{"Student", "Cho", "Chang", "cho.chang@quality.firecloud.org"},
			{"Student", "Oliver", "Wood", "oliver.wood@quality.firecloud.org"},
			{"Student", "Cedric", "Diggory", "cedric.diggory@quality.firecloud.org"},
			{"Student", "Vincent", "Crabbe", "vincent.crabbe@quality.firecloud.org"},
			{"Student", "Gregory", "Goyle", "gregory.goyle@quality.firecloud.org"},
			{"Student", "Dean", "Thomas", "dean.thomas@quality.firecloud.org"},
			{"Student", "Ginny", "Weasley", "ginny.weasley@quality.firecloud.org"},

			// Auth Domain Users
			{"Researcher", "Fred", "Weasley", "fred.authdomain@quality.firecloud.org"},
			{"Researcher", "George", "Weasley", "george.authdomain@quality.firecloud.org"},
			{"Researcher", "Bill", "Weasley", "bill.authdomain@quality.firecloud.org"},
			{"Researcher", "Molly", "Weasley", "molly.authdomain@quality.firecloud.org"},
			{"Researcher", "Arthur", "Weasley", "arthur.authdomain@quality.firecloud.org"},
			{"Researcher", "Percy", "Weasley", "percy.authdomain@quality.firecloud.org"},
		}
	}
	return config, nil
}

func (s *seeder) configWithAgoraData() (seedConfig, error) {
	config, err := s.configWithBasicDefaults()
	if err != nil {
		return config, err
	}
	if len(config.Agora.Methods) > 0 {
		log.Info().Msg("agora methods already present in config, skipping defaults")
	} else {
		config.Agora.Methods = []AgoraPayload{
			{
				Namespace:     "automationmethods",
				Name:          "DO_NOT_CHANGE_test_method",
				Synopsis:      "testtestsynopsis",
				Documentation: "",
				EntityType:    "Workflow",
				Payload: strings.TrimSpace(`
task hello {
  String? name

  command {
    echo 'hello ${name}!'
  }
  output {
    File response = stdout()
  }
  runtime {
    docker: "ubuntu"
  }
}

workflow test {
  call hello
}
`),
			},
			{
				Namespace:     "automationmethods",
				Name:          "DO_NOT_CHANGE_test_method_input_required",
				Synopsis:      "method with required inputs for testing",
				Documentation: "",
				EntityType:    "Workflow",
				Payload: strings.TrimSpace(`
task hello {
  String name

  command {
    echo 'hello ${name}!'
  }
  output {
    File response = stdout()
  }
  runtime {
    docker: "ubuntu"
  }
}

workflow test {
  call hello
}
`),
			},
		}
	}
	if len(config.Agora.Configurations) > 0 {
		log.Info().Msg("agora configurations already present in config, skipping defaults")
	} else {
		config.Agora.Configurations = []AgoraPayload{
			{
				Namespace:     "automationmethods",
				Name:          "DO_NOT_CHANGE_test1_config",
				Synopsis:      "",
				Documentation: "",
				EntityType:    "Configuration",
				Payload: strings.TrimSpace(`
{
  "name": "DO_NOT_CHANGE_test1_config",
  "methodRepoMethod": {
    "methodNamespace": "automationmethods",
    "methodName": "DO_NOT_CHANGE_test_method",
    "methodVersion": 1
  },
  "outputs": {
    "test.hello.response": "this.output"
  },
  "inputs": {
    "test.hello.name": "this.name"
  },
  "rootEntityType": "participant",
  "prerequisites": {},
  "namespace": "automationmethods"
}`),
			},
		}
	}
	// No defaults permissions for dev, Hermione already gets owner because she creates everything
	if len(config.Agora.Permissions.QA) > 0 {
		log.Info().Msg("agora permissions for QA already present in config, skipping defaults")
	} else {
		config.Agora.Permissions.QA = []AgoraPermission{
			{
				User: "tonks.owner@quality.firecloud.org",
				Role: "OWNER",
			},
			{
				User: "sirius.owner@quality.firecloud.org",
				Role: "OWNER",
			},
			{
				User: "hermione.owner@quality.firecloud.org", // note - rawls and orch tests depend on Hermione being an owner
				Role: "OWNER",
			},
		}
	}
	return config, nil
}
