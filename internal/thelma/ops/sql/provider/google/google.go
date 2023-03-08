package google

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/sqladmin/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"time"
)

// suffix to remove when converting
// https://cloud.google.com/sql/docs/postgres/add-manage-iam-users#grant-db-privileges
const googleIAMAuthStripSuffix = ".gserviceaccount.com"
const cloudSqlIamAuthenticationFlag = "cloudsql.iam_authentication"
const cloudSqlFlagEnabled = "on"

const operationFinishedStatus = "DONE"
const operationPollInterval = 3 * time.Second

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

func New(connection api.Connection, sqladminClient *sqladmin.Service, vaultClient *vaultapi.Client) provider.Provider {
	return &google{
		connection: connection,
		client:     sqladminClient,
		features: lazy.NewLazyE(func() (features, error) {
			instance, err := sqladminClient.Instances.Get(connection.GoogleInstance.Project, connection.GoogleInstance.InstanceName).Do()
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
	client        *sqladmin.Service
	features      lazy.LazyE[features]
	passwordStore passwordStore
	roUser        googleSA
	rwUser        googleSA
}

func (g *google) ClientSettings() (dbms.ClientSettings, error) {
	var creds *api.Credentials
	var nickname string
	var err error

	switch g.connection.Options.PermissionLevel {
	case api.ReadOnly:
		creds, err = g.getLocalThelmaUserCredentials(g.roUser)
		nickname = g.roUser.nickname()
	case api.ReadWrite:
		creds, err = g.getLocalThelmaUserCredentials(g.rwUser)
		nickname = g.rwUser.nickname()
	case api.Admin:
		creds, err = g.getLocalAdminCredentials()
	default:
		panic(fmt.Errorf("unsupported permission level: %#v", g.connection.Options.PermissionLevel))
	}

	if err != nil {
		return dbms.ClientSettings{}, err
	}

	return dbms.ClientSettings{
		Username: creds.Username,
		Password: creds.Password,
		Host:     cloudsqlProxySidecarAddress,
		Database: g.connection.Options.Database,
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
	instance, err := g.client.Instances.Get(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName).Do()
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

func (g *google) PodSpec() (podrun.ProviderSpec, error) {
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
			//"--http-address=0.0.0.0",
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
	if f.iamSupported && g.connection.Options.PermissionLevel != api.Admin {
		// use IAM login for users that are not the local admin account
		sidecar.Command = append(sidecar.Command, "-enable_iam_login")
	}
	return podrun.ProviderSpec{
		Sidecar:        &sidecar,
		ServiceAccount: g.kubernetesServiceAccount(),
	}, nil
}

func (g *google) getPasswordIfIAMNotSupported(user string) (string, error) {
	f, err := g.features.Get()
	if err != nil {
		return "", err
	}

	if f.iamSupported {
		return "", nil
	}
	return g.passwordStore.fetch(g.connection.GoogleInstance, user)
}

func (g *google) enableIAMAuthIfSupported(features features) error {
	if !features.iamSupported {
		return nil
	}
	if features.iamEnabled {
		return nil
	}

	log.Info().Msgf("Enabling Cloud IAM Authentication for %s", g.connection.GoogleInstance.InstanceName)

	patch := &sqladmin.DatabaseInstance{}
	patch.Settings = &sqladmin.Settings{
		DatabaseFlags: []*sqladmin.DatabaseFlags{
			{
				Name:  cloudSqlIamAuthenticationFlag,
				Value: cloudSqlFlagEnabled,
			},
		},
	}

	op, err := g.client.Instances.Patch(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName, patch).Do()
	if err != nil {
		return err
	}
	if err = g.waitForOpToBeDone(op); err != nil {
		return err
	}
	return nil
}

func (g *google) getLocalUsernames() (set.Set[string], error) {
	userResp, err := g.client.Users.List(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName).Do()
	if err != nil {
		return nil, err
	}

	userSet := set.NewSet[string]()
	for _, user := range userResp.Items {
		userSet.Add(user.Name)
	}

	return userSet, nil
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
			if _, err = g.client.Users.Delete(
				g.connection.GoogleInstance.Project,
				g.connection.GoogleInstance.InstanceName,
			).Name(username).Do(); err != nil {
				return err
			}
		}
	}

	for _, username := range usernames {
		log.Info().Msgf("Adding user %s", username)
		user := &sqladmin.User{
			Name: username,
		}
		if features.iamSupported {
			user.Type = "CLOUD_IAM_SERVICE_ACCOUNT"
		} else {
			user.Type = "BUILT_IN"
			password := g.passwordStore.generate()
			if err = g.passwordStore.save(g.connection.GoogleInstance, username, password); err != nil {
				return fmt.Errorf("error saving generated password to Vault: %v", err)
			}
			user.Password = password
		}

		if _, err = g.client.Users.Insert(
			g.connection.GoogleInstance.Project,
			g.connection.GoogleInstance.InstanceName,
			user,
		).Do(); err != nil {
			return err
		}
	}

	return nil
}

func (g *google) resetPassword(username string) (string, error) {
	log.Info().Msgf("Resetting password for user %s", username)

	password := g.passwordStore.generate()

	req := g.client.Users.Update(g.connection.GoogleInstance.Project, g.connection.GoogleInstance.InstanceName, &sqladmin.User{
		Password: password,
	}).Name(username)

	op, err := req.Do()
	if err != nil {
		return "", fmt.Errorf("error resetting password for %s: %v", username, err)
	}

	if err = g.waitForOpToBeDone(op); err != nil {
		return "", fmt.Errorf("error resetting password for %s: %v", username, err)
	}

	return password, nil
}

func (g *google) waitForOpToBeDone(op *sqladmin.Operation) error {
	var err error

	for op.Status != operationFinishedStatus { // TODO constant
		log.Info().Msgf("Waiting for %s operation to complete...", op.OperationType)
		op, err = g.client.Operations.Get(op.TargetProject, op.Name).Do()
		if err != nil {
			return err
		}
		time.Sleep(operationPollInterval)
	}
	return nil
}

func (g *google) databaseUser() (string, error) {
	switch g.connection.Options.PermissionLevel {
	case api.ReadOnly:
		return g.roUser.dbuser(), nil
	case api.ReadWrite:
		return g.roUser.dbuser(), nil
	case api.Admin:
		f, err := g.features.Get()
		if err != nil {
			return "", err
		}
		return f.dbms.AdminUser(), nil
	default:
		panic(fmt.Errorf("unsupported permission level: %#v", g.connection.Options.PermissionLevel))
	}
}

func (g *google) kubernetesServiceAccount() string {
	switch g.connection.Options.PermissionLevel {
	case api.ReadOnly:
		return googleSATemplates.readOnly.kubernetesSA
	case api.ReadWrite:
		return googleSATemplates.readWrite.kubernetesSA
	case api.Admin:
		// admin level uses password auth, not iam, so the SA doesn't actually matter
		return googleSATemplates.readOnly.kubernetesSA
	default:
		panic(fmt.Errorf("unsupported permission level: %#v", g.connection.Options.PermissionLevel))
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
