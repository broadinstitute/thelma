package kubernetes

import (
	"context"
	"fmt"
	k8s "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/provider"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils/lazy"
	"github.com/broadinstitute/thelma/internal/thelma/utils/maps"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pwgen"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
)

const readonlyUsername = "thelma-sql-ro"
const readwriteUsername = "thelma-sql-rw"
const serviceAccountName = "thelma-workloads"
const secretNameSuffix = "-thelma-sql-secret"

func New(connection api.Connection, clients k8s.Clients) (provider.Provider, error) {
	return newKubernetesProvider(connection, clients, pwgen.Pwgen{MinLength: 8})
}

// package-private constructor for testing
func newKubernetesProvider(connection api.Connection, clients k8s.Clients, pwg pwgen.Generator) (provider.Provider, error) {
	_kubecfg, err := clients.Kubecfg()
	if err != nil {
		return nil, err
	}
	_kubectx, err := _kubecfg.ForRelease(connection.KubernetesInstance.Release)
	if err != nil {
		return nil, err
	}
	_k8sclient, err := clients.ForKubectx(_kubectx)
	if err != nil {
		return nil, err
	}
	_kubectl, err := clients.Kubectl()
	if err != nil {
		return nil, err
	}

	return &kubernetes{
		conn:    connection,
		kubectx: _kubectx,
		kubectl: _kubectl,
		client:  _k8sclient,
		features: lazy.NewLazyE(func() (*features, error) {
			return detectFeatures(connection.KubernetesInstance.Release, _k8sclient)
		}),
		pwg: pwg,
	}, nil
}

type kubernetes struct {
	conn     api.Connection
	kubectx  kubecfg.Kubectx
	client   k8sclient.Interface
	kubectl  kubectl.Kubectl
	features lazy.LazyE[*features]
	pwg      pwgen.Generator
}

func (k *kubernetes) ClientSettings(overrides ...provider.ConnectionOverride) (dbms.ClientSettings, error) {
	options := k.conn.Options
	for _, o := range overrides {
		o(&options)
	}

	f, err := k.features.Get()
	if err != nil {
		return dbms.ClientSettings{}, err
	}

	creds, err := k.getCredentialsForConnection(options)
	if err != nil {
		return dbms.ClientSettings{}, err
	}

	roUser, err := k.getLocalThelmaUserCredentials(readonlyUsername)
	if err != nil {
		return dbms.ClientSettings{}, err
	}
	rwUser, err := k.getLocalThelmaUserCredentials(readwriteUsername)
	if err != nil {
		return dbms.ClientSettings{}, err
	}

	return dbms.ClientSettings{
		Username: creds.Username,
		Password: creds.Password,
		Host:     f.serviceHostName,
		Database: options.Database,
		Init: dbms.InitSettings{
			CreateUsers: true,
			ReadOnlyUser: dbms.InitUser{
				Name:     roUser.Username,
				Password: roUser.Password,
			},
			ReadWriteUser: dbms.InitUser{
				Name:     rwUser.Username,
				Password: rwUser.Password,
			},
		},
	}, nil
}

func (k *kubernetes) DetectDBMS() (api.DBMS, error) {
	f, err := k.features.Get()
	if err != nil {
		return -1, err
	}
	return f.dbms, nil
}

func (k *kubernetes) Initialized() (bool, error) {
	return k.secretExists()
}

func (k *kubernetes) Initialize() error {
	exists, err := k.secretExists()
	if err != nil {
		return err
	}
	if exists {
		if err = k.client.CoreV1().Secrets(k.namespace()).
			Delete(context.Background(), k.secretName(), metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: k.secretName(),
		},
		StringData: map[string]string{
			readonlyUsername:  k.pwg.Generate(),
			readwriteUsername: k.pwg.Generate(),
		},
	}
	_, err = k.client.CoreV1().Secrets(k.namespace()).Create(context.Background(), &secret, metav1.CreateOptions{})
	return err
}

func (k *kubernetes) PodSpec(_ ...provider.ConnectionOverride) (podrun.ProviderSpec, error) {
	return podrun.ProviderSpec{
		Sidecar:        nil,
		ServiceAccount: serviceAccountName,
	}, nil
}

func (k *kubernetes) getCredentialsForConnection(options api.ConnectionOptions) (*api.Credentials, error) {
	switch options.PrivilegeLevel {
	case api.Admin:
		return k.getAdminUserCredentials()
	case api.ReadWrite:
		return k.getLocalThelmaUserCredentials(readwriteUsername)
	case api.ReadOnly:
		return k.getLocalThelmaUserCredentials(readonlyUsername)
	default:
		panic(errors.Errorf("unsupported permission level: %#v", options.PrivilegeLevel))
	}
}

func (k *kubernetes) resetPassword(username string) (*api.Credentials, error) {
	f, err := k.features.Get()
	if err != nil {
		return nil, err
	}

	password := k.pwg.Generate()
	log.Info().Msgf("Resetting password for user %s", username)
	var command []string
	switch f.dbms {
	case api.Postgres:
		command = []string{"psql", "--no-psqlrc", "--host", "localhost", "-U", api.Postgres.AdminUser(), "-c", fmt.Sprintf("alter user %s password '%s';", username, password)}
	case api.MySQL:
		panic("TODO")
	default:
		panic(errors.Errorf("unsupported dbms type: %#v", f.dbms))
	}

	if err = k.kubectl.Exec(k.kubectx, f.container, command); err != nil {
		return nil, errors.Errorf("error resetting password for %s: %v", username, err)
	}

	return &api.Credentials{
		Username: username,
		Password: password,
	}, nil
}

func (k *kubernetes) getAdminUserCredentials() (*api.Credentials, error) {
	f, err := k.features.Get()
	if err != nil {
		return nil, err
	}

	return k.resetPassword(f.dbms.AdminUser())
}

func (k *kubernetes) getLocalThelmaUserCredentials(username string) (*api.Credentials, error) {
	s, err := k.readSecret()
	if err != nil {
		return nil, err
	}
	password, exists := s[username]
	if !exists {
		return nil, errors.Errorf("no password found for %s in secret %s", username, k.secretName())
	}
	return &api.Credentials{
		Username: username,
		Password: password,
	}, nil
}

func (k *kubernetes) readSecret() (map[string]string, error) {
	s, err := k.client.CoreV1().Secrets(k.namespace()).Get(context.Background(), k.secretName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return maps.TransformValues(s.Data, func(v []byte) string {
		return string(v)
	}), nil
}

func (k *kubernetes) secretExists() (bool, error) {
	// we use list here because it doesn't seem like there's a great
	// way to distinguish between 404 (secret not found) and other
	// types of API errors (5xx/4xx) received from cluster API
	secrets, err := k.client.CoreV1().Secrets(k.namespace()).List(
		context.Background(),
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", k.secretName()),
		},
	)
	if err != nil {
		return false, err
	}
	if len(secrets.Items) == 0 {
		return false, nil
	}
	if len(secrets.Items) == 1 {
		return true, nil
	}
	// multiple secrets with the same name in the same namespace should be impossible
	panic(errors.Errorf("found multiple secrets with name %s (%d)", k.secretName(), len(secrets.Items)))
}

func (k *kubernetes) secretName() string {
	return secretName(k.release())
}

func (k *kubernetes) namespace() string {
	return k.release().Namespace()
}

func (k *kubernetes) release() terra.Release {
	return k.conn.KubernetesInstance.Release
}

func secretName(r terra.Release) string {
	return r.Name() + secretNameSuffix
}
