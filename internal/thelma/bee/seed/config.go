package seed

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
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
			VaultPath string `default:"secret/dsde/firecloud/%s/rawls/rawls-account.json"`
			VaultKey  string `default:""`
		}
		Sam struct {
			VaultPath string `default:"secret/dsde/firecloud/%s/sam/sam-account.json"`
			VaultKey  string `default:""`
		}
		Leonardo struct {
			VaultPath string `default:"secret/dsde/firecloud/%s/leonardo/leonardo-account.json"`
			VaultKey  string `default:""`
		}
		FirecloudOrch struct {
			VaultPath string `default:"secret/dsde/firecloud/%s/common/firecloud-account.json"`
			VaultKey  string `default:""`
		}
		WorkspaceManager struct {
			// WSM dev SA used for both Dev and QA BEEs, as of 7/13/2022
			VaultPath string `default:"secret/dsde/terra/kernel/dev/dev/workspace/app-sa"`
			VaultKey  string `default:"key.json"`
		}
		TSPS struct {
			VaultPath string `default:"secret/dsde/firecloud/%s/tsps/tsps-account.json"`
			VaultKey  string `default:""`
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
	var vaultPath, vaultKey string
	switch appRelease.Name() {
	case "rawls":
		vaultPath = config.Auth.Rawls.VaultPath
		vaultKey = config.Auth.Rawls.VaultKey
	case "sam":
		vaultPath = config.Auth.Sam.VaultPath
		vaultKey = config.Auth.Sam.VaultKey
	case "leonardo":
		vaultPath = config.Auth.Leonardo.VaultPath
		vaultKey = config.Auth.Leonardo.VaultKey
	case "firecloudorch":
		vaultPath = config.Auth.FirecloudOrch.VaultPath
		vaultKey = config.Auth.FirecloudOrch.VaultKey
	case "workspacemanager":
		vaultPath = config.Auth.WorkspaceManager.VaultPath
		vaultKey = config.Auth.WorkspaceManager.VaultKey
	case "tsps":
		vaultPath = config.Auth.TSPS.VaultPath
		vaultKey = config.Auth.TSPS.VaultKey
	default:
		return nil, errors.Errorf("thelma doesn't know how to authenticate as %s", appRelease.Name())
	}
	if strings.ContainsRune(vaultPath, '%') {
		vaultPath = fmt.Sprintf(vaultPath, appRelease.Cluster().ProjectSuffix())
	}
	return s.clientFactory.Google(
		append(options, google.OptionForceVaultSA(vaultPath, vaultKey))...,
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
