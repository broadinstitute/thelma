package google

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/utils/pwgen"
	vaultapi "github.com/hashicorp/vault/api"
	"path"
	"time"
)

const nonprodPrefix = "secret/dsde/terra/thelma/sql/google"
const prodPrefix = "secret/suitable/terra/thelma/sql/google"
const passwordSecretKey = "password"

type passwordStore interface {
	generate() string
	save(instance api.GoogleInstance, username string, password string) error
	fetch(instance api.GoogleInstance, username string) (string, error)
}

func newPasswordStore(vaultClient *vaultapi.Client) passwordStore {
	return &passwordStoreImpl{
		vaultClient: vaultClient,
		generator: pwgen.Pwgen{
			MinLength:  12,
			MinLower:   1,
			MinUpper:   1,
			MinNum:     1,
			MinSpecial: 1,
		},
	}
}

type passwordStoreImpl struct {
	generator   pwgen.Pwgen
	vaultClient *vaultapi.Client
}

func (p *passwordStoreImpl) generate() string {
	return p.generator.Generate()
}

func (p *passwordStoreImpl) save(instance api.GoogleInstance, username string, password string) error {
	_, err := p.vaultClient.Logical().Write(p.secretPath(instance, username), map[string]interface{}{
		passwordSecretKey: password,
		"comment":         fmt.Sprintf("Generated at %s by thelma sql connect", time.Now().Format(time.RFC822)),
	})
	return err
}

func (p *passwordStoreImpl) fetch(instance api.GoogleInstance, username string) (string, error) {
	secretPath := p.secretPath(instance, username)

	secret, err := p.vaultClient.Logical().Read(secretPath)
	if err != nil {
		return "", err
	}
	password, exists := secret.Data[passwordSecretKey]
	if !exists {
		return "", fmt.Errorf("vault secret at %s has no %s field; try re-running thelma sql init for this database", secretPath, passwordSecretKey)
	}
	s, ok := password.(string)
	if !ok {
		return "", fmt.Errorf("vault secret at %s has unexpected value type for field %s (want string); try re-running thelma sql init for this database", secretPath, passwordSecretKey)
	}
	return s, nil
}

func (p *passwordStoreImpl) secretPath(instance api.GoogleInstance, username string) string {
	prefix := nonprodPrefix
	if instance.IsProd() {
		prefix = prodPrefix
	}
	return path.Join(prefix, instance.Project, instance.InstanceName, username)
}
