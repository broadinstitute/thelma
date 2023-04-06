package sqladmin

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/sqladmin/v1"
	"time"
)

const operationFinishedStatus = "DONE"
const operationPollInterval = 3 * time.Second

// Client a wrapper for the go-generated sqladmin api client library
// we do this so that we can generate mocks to test code that makes api
// calls (unlike other GCP APIs, the sqladmin client is a struct instead of
// an interface).
type Client interface {
	GetInstance(project string, instanceName string) (*sqladmin.DatabaseInstance, error)
	PatchInstance(project string, instanceName string, patchRequest *sqladmin.DatabaseInstance) error
	GetInstanceLocalUsers(project string, instanceName string) ([]string, error)
	ResetPassword(project string, instanceName string, username string, password string) error
	DeleteUser(project string, instanceName string, username string) error
	AddUser(project string, instanceName string, user *sqladmin.User) error
}

func New(sqladminClient *sqladmin.Service) Client {
	return &client{
		sqladminClient: sqladminClient,
	}
}

type client struct {
	sqladminClient *sqladmin.Service
}

func (c client) GetInstance(project string, instanceName string) (*sqladmin.DatabaseInstance, error) {
	return c.sqladminClient.Instances.Get(project, instanceName).Do()
}

func (c client) PatchInstance(project string, instanceName string, patchRequest *sqladmin.DatabaseInstance) error {
	op, err := c.sqladminClient.Instances.Patch(project, instanceName, patchRequest).Do()
	if err != nil {
		return err
	}
	if err = c.waitForOpToBeDone(op); err != nil {
		return err
	}
	return nil
}

func (c client) GetInstanceLocalUsers(project string, instanceName string) ([]string, error) {
	userResp, err := c.sqladminClient.Users.List(project, instanceName).Do()
	if err != nil {
		return nil, err
	}

	var users []string
	for _, user := range userResp.Items {
		users = append(users, user.Name)
	}

	return users, nil
}

func (c client) ResetPassword(project string, instanceName string, username string, password string) error {
	req := c.sqladminClient.Users.Update(project, instanceName, &sqladmin.User{
		Password: password,
	}).Name(username)

	op, err := req.Do()
	if err != nil {
		return fmt.Errorf("error resetting password for %s user %s: %v", instanceName, username, err)
	}

	if err = c.waitForOpToBeDone(op); err != nil {
		return fmt.Errorf("error resetting password for %s: %v", username, err)
	}

	return nil
}

func (c client) DeleteUser(project string, instanceName string, username string) error {
	op, err := c.sqladminClient.Users.Delete(
		project, instanceName,
	).Name(username).Do()
	if err != nil {
		return fmt.Errorf("error deleting user %s from instance %s: %v", username, instanceName, err)
	}

	return c.waitForOpToBeDone(op)
}

func (c client) AddUser(project string, instanceName string, user *sqladmin.User) error {
	op, err := c.sqladminClient.Users.Insert(
		project, instanceName, user,
	).Do()
	if err != nil {
		return fmt.Errorf("error adding user %s to instance %s: %v", user.Name, instanceName, err)
	}

	return c.waitForOpToBeDone(op)
}

func (c client) waitForOpToBeDone(op *sqladmin.Operation) error {
	var err error

	for op.Status != operationFinishedStatus {
		log.Info().Msgf("Waiting for operation to complete...")
		op, err = c.sqladminClient.Operations.Get(op.TargetProject, op.Name).Do()
		if err != nil {
			return err
		}
		time.Sleep(operationPollInterval)
	}
	return nil
}
