// Package cleanup contains logic for cleaning up Bee cloud resources when BEEs are deleted
package cleanup

import (
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/rs/zerolog/log"
)

// topicIdFormats list of pubsub topics created by services running in BEEs that should be deleted
// when the BEE is deleted (%s is substituted with environment name).
// I HATE HATE HATE that we need to keep this list, but the go pubsub client apparently has no way of filtering
// topic ids by regular expression. If we wanted to just match by environment name, we'd be iterating
// through thousands of topic names client-side. :skull:
var topicIdFormats = []string{
	"leonardo-pubsub-%s",
	"rawls-async-import-topic-%s",
	"sam-group-sync-%s",
	"terra-%s-stairwaycluster-workqueue",
	"workbench-notifications-%s",
}

type Cleanup interface {
	Cleanup(bee terra.Environment) error
}

func NewCleanup(googleClients google.Clients) Cleanup {
	return &cleanup{
		googleClients: googleClients,
	}
}

type cleanup struct {
	googleClients google.Clients
}

func (c *cleanup) Cleanup(bee terra.Environment) error {
	if !bee.Lifecycle().IsDynamic() {
		panic(fmt.Errorf("%s is not a dynamic environment, won't attempt to cleanup its resources", bee.Name()))
	}

	return c.cleanupPubsubTopics(bee)
}

func (c *cleanup) cleanupPubsubTopics(env terra.Environment) error {
	projects := projectIds(env)
	for _, projectId := range projects {
		if err := c.cleanupPubsubTopicsInProject(env, projectId); err != nil {
			return err
		}
	}
	return nil
}

// clean up pubsub topics in the project
func (c *cleanup) cleanupPubsubTopicsInProject(env terra.Environment, projectId string) error {
	client, err := c.googleClients.PubSub(projectId)
	if err != nil {
		return err
	}

	for _, topicId := range pubsubTopicIds(env) {
		if err = client.Topic(topicId).Delete(context.Background()); err != nil {
			// If the environment was never created successfully, the pubsub topic might not exist, so just log a warning and exit
			log.Warn().Err(err).Msgf("Failed to delete pubsub topic %s in project %s", topicId, projectId)
		} else {
			log.Info().Msgf("Deleted pubsub topic %s in project %s", topicId, projectId)
		}
	}

	return nil
}

// return the list of topic ids that should be cleaned up for this environment
func pubsubTopicIds(env terra.Environment) []string {
	topicIds := set.NewStringSet()

	for _, topicId := range topicIdFormats {
		topicIds.Add(fmt.Sprintf(topicId, env.Name()))
	}

	return topicIds.Elements()
}

// return the list of projects this terra.Environment is distributed across
// in 99% of cases this will be a single project
func projectIds(env terra.Environment) []string {
	projects := set.NewStringSet()
	for _, r := range env.Releases() {
		projects.Add(r.Cluster().Project())
	}
	return projects.Elements()
}
