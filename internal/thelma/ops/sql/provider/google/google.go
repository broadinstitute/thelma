package google

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/sqladmin"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	googlesqladmin "google.golang.org/api/sqladmin/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

// suffix to remove when converting
// https://cloud.google.com/sql/docs/postgres/add-manage-iam-users#grant-db-privileges
const googleIAMAuthStripSuffix = ".gserviceaccount.com"
const cloudSqlIamAuthenticationFlag = "cloudsql.iam_authentication"
const cloudSqlFlagEnabled = "on"
const cloudSqlAccountTypeIAM = "CLOUD_IAM_SERVICE_ACCOUNT"
const cloudSqlAccountTypeBuiltIn = "BUILT_IN"

const cloudsqlProxySidecarAddress = "127.0.0.1"

// googleSATemplates holds google SAs for different connection permission levels
var googleSATemplates = struct {
	readOnly  googleSATemplate
	readWrite googleSATemplate
}{
	readOnly: googleSATemplate{
		emailTemplate: "thelma-sql-ro-<CLUSTER>@<PROJECT>.iam.gserviceaccount.com",
		kubernetesSA:  "thelma-sql-ro",
	},
	readWrite: googleSATemplate{
		emailTemplate: "thelma-sql-rw-<CLUSTER>@<PROJECT>.iam.gserviceaccount.com",
		kubernetesSA:  "thelma-sql-rw",
	},
}

func New(connection api.Connection, sqladminClient sqladmin.Client, vaultClient *vaultapi.Client) provider.Provider {
	return &google{
		connection: connection,
		client:     sqladminClient,
		features: lazy.NewLazyE(func() (features, error) {
			instance, err := sqladminClient.GetInstance(connection.GoogleInstance.Project, connection.GoogleInstance.InstanceName)
			if err != nil {
				return features{}, err
			}
			return extractFeatures(instance)
		}),
		passwordStore: newPasswordStore(vaultClient),
		roUser: googleSA{
			googleSATemplate: googleSATemplates.readOnly,
			conn:             connection,
		},
		rwUser: googleSA{
			googleSATemplate: googleSATemplates.readWrite,
			conn:             connection,
		},
	}
}

type google struct {
	connection    api.Connection
	client        sqladmin.Client
	features      lazy.LazyE[features]
	passwordStore passwordStore
	roUser        googleSA
	rwUser        googleSA
}

func (g *google) ClientSettings(overrides ...provider.ConnectionOverride) (dbms.ClientSettings, error) {
	options := g.connection.Options
	for _, override := range overrides {
		override(&options)
	}

	var creds *api.Credentials
	var nickname string
	var err error

	switch options.PrivilegeLevel {
	case api.ReadOnly:
		creds, err = g.getLocalThelmaUserCredentials(g.roUser)
		nickname = g.roUser.nickname()
	case api.ReadWrite:
		creds, err = g.getLocalThelmaUserCredentials(g.rwUser)
		nickname = g.rwUser.nickname()
	case api.Admin:
		creds, err = g.getLocalAdminCredentials()
	default:
		panic(fmt.Errorf("unsupported permission level: %#v", g.connection.Options.PrivilegeLevel))
	}

	if err != nil {
		return dbms.ClientSettings{}, err
	}

	return dbms.ClientSettings{
		Username: creds.Username,
		Password: creds.Password,
		Host:     cloudsqlProxySidecarAddress,
		Database: options.Database,
		Nickname: nickname,
		Init: dbms.InitSettings{
			CreateUsers: false, // we created them already via CloudSQL API commands
			ReadOnlyUser: dbms.InitUser{
				Name: g.roUser.dbuser(),
			},
			ReadWriteUser: dbms.InitUser{
				Name: g.rwUser.dbuser(),
			},
		},
	}, nil
}

func (g *google) DetectDBMS() (api.DBMS, error) {
	f, err := g.features.Get()
	if err != nil {
		return -1, err
	}
	return f.dbms, nil
}

func (g *google) Initialized() (bool, error) {
	userSet, err := g.getLocalUsernames()
	if err != nil {
		return false, err
	}

	return userSet.Exists(g.roUser.dbuser()) && userSet.Exists(g.rwUser.dbuser()), nil
}

func (g *google) Initialize() error {
	instance, err := g.client.GetInstance(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName)
	if err != nil {
		return err
	}

	feats, err := extractFeatures(instance)
	if err != nil {
		return err
	}

	if err = g.enableIAMAuthIfSupported(feats); err != nil {
		return err
	}

	if err = g.addThelmaUsers(feats); err != nil {
		return err
	}

	return nil
}

func (g *google) PodSpec(overrides ...provider.ConnectionOverride) (podrun.ProviderSpec, error) {
	options := g.connection.Options
	for _, override := range overrides {
		override(&options)
	}

	_features, err := g.features.Get()
	if err != nil {
		return podrun.ProviderSpec{}, err
	}

	sidecar := v1.Container{
		Name:  "sqlproxy",
		Image: "gcr.io/cloudsql-docker/gce-proxy:latest",
		Command: []string{
			"/cloud_sql_proxy",
			"-instances=$(SQL_INSTANCE_PROJECT):$(SQL_INSTANCE_REGION):$(SQL_INSTANCE_NAME)=tcp:5432",
			"-use_http_health_check",
			"-health_check_port=9090",
			"-verbose",
		},
		Env: []v1.EnvVar{
			{
				Name:  "SQL_INSTANCE_PROJECT",
				Value: g.connection.GoogleInstance.Project,
			},
			{
				Name:  "SQL_INSTANCE_REGION",
				Value: _features.region,
			}, {
				Name:  "SQL_INSTANCE_NAME",
				Value: g.connection.GoogleInstance.InstanceName,
			},
		},
		LivenessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/liveness",
					Port: intstr.FromInt(9090),
				},
			},
			PeriodSeconds:    60,
			TimeoutSeconds:   30,
			FailureThreshold: 5,
		},
		ReadinessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/readiness",
					Port: intstr.FromInt(9090),
				},
			},
			PeriodSeconds:    10,
			TimeoutSeconds:   5,
			SuccessThreshold: 1,
			FailureThreshold: 1,
		},
		StartupProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: "/startup",
					Port: intstr.FromInt(9090),
				},
			},
			InitialDelaySeconds: 0,
			TimeoutSeconds:      5,
			PeriodSeconds:       1,
			SuccessThreshold:    0,
			FailureThreshold:    20,
		},
	}

	f, err := g.features.Get()
	if err != nil {
		return podrun.ProviderSpec{}, err
	}
	if f.iamSupported && options.PrivilegeLevel != api.Admin {
		// use IAM login for users that are not the local admin account
		sidecar.Command = append(sidecar.Command, "-enable_iam_login")
	}
	return podrun.ProviderSpec{
		Sidecar:        &sidecar,
		ServiceAccount: g.kubernetesServiceAccount(),
	}, nil
}

func (g *google) enableIAMAuthIfSupported(features features) error {
	if !features.iamSupported {
		return nil
	}
	if features.iamEnabled {
		return nil
	}

	log.Info().Msgf("Enabling Cloud IAM Authentication for %s", g.connection.GoogleInstance.InstanceName)

	patch := &googlesqladmin.DatabaseInstance{}
	patch.Settings = &googlesqladmin.Settings{
		DatabaseFlags: []*googlesqladmin.DatabaseFlags{
			{
				Name:  cloudSqlIamAuthenticationFlag,
				Value: cloudSqlFlagEnabled,
			},
		},
	}

	return g.client.PatchInstance(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName, patch)
}

func (g *google) getLocalUsernames() (set.Set[string], error) {
	users, err := g.client.GetInstanceLocalUsers(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName)
	if err != nil {
		return nil, err
	}
	return set.NewSet[string](users...), nil
}

func (g *google) getLocalAdminCredentials() (*api.Credentials, error) {
	f, err := g.features.Get()
	if err != nil {
		return nil, err
	}

	username := f.dbms.AdminUser()
	password, err := g.resetPassword(username)
	if err != nil {
		return nil, err
	}
	return &api.Credentials{
		Username: username,
		Password: password,
	}, nil
}

func (g *google) getLocalThelmaUserCredentials(user googleSA) (*api.Credentials, error) {
	f, err := g.features.Get()
	if err != nil {
		return nil, err
	}

	if f.iamSupported {
		return &api.Credentials{
			Username: user.dbuser(),
			Password: "",
		}, nil
	}

	password, err := g.passwordStore.fetch(g.connection.GoogleInstance, user.dbuser())
	if err != nil {
		return nil, err
	}
	return &api.Credentials{
		Username: user.dbuser(),
		Password: password,
	}, nil
}

func (g *google) addThelmaUsers(features features) error {
	var err error

	userSet, err := g.getLocalUsernames()
	if err != nil {
		return err
	}

	usernames := []string{
		g.roUser.dbuser(),
		g.rwUser.dbuser(),
	}

	for _, username := range usernames {
		if userSet.Exists(username) {
			log.Info().Msgf("Deleting existing user %s (will re-create)", username)
			if err = g.client.DeleteUser(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName, username); err != nil {
				return err
			}
		}
	}

	for _, username := range usernames {
		log.Info().Msgf("Adding user %s", username)
		user := &googlesqladmin.User{
			Name: username,
		}
		if features.iamSupported {
			user.Type = cloudSqlAccountTypeIAM
		} else {
			user.Type = cloudSqlAccountTypeBuiltIn
			password := g.passwordStore.generate()
			if err = g.passwordStore.save(g.connection.GoogleInstance, username, password); err != nil {
				return fmt.Errorf("error saving generated password to Vault: %v", err)
			}
			user.Password = password
		}

		if err = g.client.AddUser(
			g.connection.GoogleInstance.Project,
			g.connection.GoogleInstance.InstanceName,
			user,
		); err != nil {
			return err
		}
	}

	return nil
}

func (g *google) resetPassword(username string) (string, error) {
	log.Info().Msgf("Resetting password for user %s", username)

	password := g.passwordStore.generate()

	err := g.client.ResetPassword(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName, username, password)
	if err != nil {
		return "", err
	}
	return password, nil
}

func (g *google) kubernetesServiceAccount() string {
	switch g.connection.Options.PrivilegeLevel {
	case api.ReadOnly:
		return googleSATemplates.readOnly.kubernetesSA
	case api.ReadWrite:
		return googleSATemplates.readWrite.kubernetesSA
	case api.Admin:
		// admin level uses password auth, not iam, so the SA doesn't actually matter
		return googleSATemplates.readOnly.kubernetesSA
	default:
		panic(fmt.Errorf("unsupported permission level: %#v", g.connection.Options.PrivilegeLevel))
	}
}

// googleSATemplate represents a template for a google service account provisioned for Thelma
type googleSATemplate struct {
	// emailTemplate for generating a google SA email for a given proxy cluster & project
	emailTemplate string
	// kubernetesSA kubernetes SA in the proxy cluster with workload identity permissions for the google SA
	kubernetesSA string
}

// googleSA represents a google service account provisioned for Thelma
type googleSA struct {
	googleSATemplate
	conn api.Connection
}

// convert a google SA email template to a local database user
// https://cloud.google.com/sql/docs/postgres/add-manage-iam-users#grant-db-privileges
// according to the above docs, this means removing the .gserviceaccount.com suffix
func (g googleSA) dbuser() string {
	email := g.googleEmail()
	return strings.TrimSuffix(email, googleIAMAuthStripSuffix)
}

func (g googleSA) nickname() string {
	return g.kubernetesSA
}

func (g googleSA) googleEmail() string {
	email := g.emailTemplate
	email = strings.ReplaceAll(email, "<CLUSTER>", g.conn.Options.ProxyCluster.Name())
	email = strings.ReplaceAll(email, "<PROJECT>", g.conn.GoogleInstance.Project)
	return email
}
