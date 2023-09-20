// Package cleanup contains logic for cleaning up Bee cloud resources when BEEs are deleted
package cleanup

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

// topicIdFormats list of pubsub topics created by services running in BEEs that should be deleted
// when the BEE is deleted (%s is substituted with environment name).
// I HATE HATE HATE that we need to keep this list in thelma, but the go pubsub client apparently has no way of filtering
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
		panic(errors.Errorf("%s is not a dynamic environment, won't attempt to cleanup its resources", bee.Name()))
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
	log.Info().Msgf("Deleting PubSub topics and subscriptions for %s in %s", env.Name(), projectId)

	for _, topicId := range pubsubTopicIds(env) {
		if err := c.deleteTopicAndSubscriptions(projectId, topicId); err != nil {
			return err
		}
	}

	return nil
}

func (c *cleanup) deleteTopicAndSubscriptions(projectId string, topicId string) error {
	client, err := c.googleClients.PubSub(projectId)
	if err != nil {
		return err
	}

	topic := client.Topic(topicId)
	exists, err := topic.Exists(context.Background())
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	// get a list of subscriptions associated with the topic
	var subs []*pubsub.Subscription
	it := topic.Subscriptions(context.Background())
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Errorf("error listing subscriptions for pubsub topic %s/%s: %v", projectId, topicId, err)
		}
		subs = append(subs, sub)
	}

	// delete subscriptions
	for _, sub := range subs {
		if err := sub.Delete(context.Background()); err != nil {
			return errors.Errorf("error deleting subscription %s for pubsub topic %s/%s: %v", sub.ID(), projectId, topicId, err)
		}
		log.Debug().Msgf("Deleted subscription: %s", sub.ID())
	}

	// delete the topic
	err = topic.Delete(context.Background())
	if err != nil {
		return errors.Errorf("error deleting pubsub topic %s/%s: %v", projectId, topicId, err)
	}

	log.Debug().Msgf("Deleted topic: %s", topicId)

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
