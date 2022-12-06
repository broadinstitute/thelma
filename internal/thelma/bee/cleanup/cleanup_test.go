package cleanup

import (
	"testing"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/mocks"
	googletesting "github.com/broadinstitute/thelma/internal/thelma/clients/google/testing"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statefixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Cleanup(t *testing.T) {
	bee := statemocks.NewEnvironment(t)
	bee.On("Name").Return("fake-bee")
	bee.On("Lifecycle").Return(terra.Dynamic)

	cluster := statemocks.NewCluster(t)
	cluster.On("Project").Return("bee-project")

	release := statemocks.NewAppRelease(t)
	release.On("Cluster").Return(cluster)

	bee.On("Releases").Return([]terra.Release{release})

	// 1 topic, 3 subscriptions
	psMocks, psClient := googletesting.NewMockPubSubServerAndClient(t, "bee-project")
	psMocks.ExpectGetTopic("leonardo-pubsub-fake-bee", &pubsubpb.Topic{
		Name: "leonardo-pubsub-fake-bee",
	}, nil)
	psMocks.ExpectListTopicSubscriptions("leonardo-pubsub-fake-bee", []string{
		"leonardo-pubsub-fake-bee-sub-1",
		"leonardo-pubsub-fake-bee-sub-2",
		"leonardo-pubsub-fake-bee-sub-3",
	}, nil)

	// pretend sam-group-sync-fake-bee exists but has no subscriptions
	psMocks.ExpectGetTopic("sam-group-sync-fake-bee", &pubsubpb.Topic{
		Name: "sam-group-sync-fake-bee",
	}, nil)
	psMocks.ExpectListTopicSubscriptions("sam-group-sync-fake-bee", []string{}, nil)

	// pretend rawls async import topic does not exist
	psMocks.ExpectGetTopic("rawls-async-import-topic-fake-bee", nil, googletesting.NotFoundError())

	// 1 topic, 1 subscription
	psMocks.ExpectGetTopic("terra-fake-bee-stairwaycluster-workqueue", &pubsubpb.Topic{
		Name: "terra-fake-bee-stairwaycluster-workqueue",
	}, nil)
	psMocks.ExpectListTopicSubscriptions("terra-fake-bee-stairwaycluster-workqueue", []string{
		"stairwaycluster-workqueue-sub-1",
	}, nil)

	// 1 topic, 1 subscription
	psMocks.ExpectGetTopic("workbench-notifications-fake-bee", &pubsubpb.Topic{
		Name: "workbench-notifications-fake-bee",
	}, nil)
	psMocks.ExpectListTopicSubscriptions("workbench-notifications-fake-bee", []string{
		"workbench-notifications-fake-bee-sub-1",
	}, nil)

	psMocks.ExpectDeleteSubscription("leonardo-pubsub-fake-bee-sub-1", nil)
	psMocks.ExpectDeleteSubscription("leonardo-pubsub-fake-bee-sub-2", nil)
	psMocks.ExpectDeleteSubscription("leonardo-pubsub-fake-bee-sub-3", nil)
	psMocks.ExpectDeleteTopic("leonardo-pubsub-fake-bee", nil)

	psMocks.ExpectDeleteTopic("sam-group-sync-fake-bee", nil)

	psMocks.ExpectDeleteSubscription("stairwaycluster-workqueue-sub-1", nil)
	psMocks.ExpectDeleteTopic("terra-fake-bee-stairwaycluster-workqueue", nil)

	psMocks.ExpectDeleteSubscription("workbench-notifications-fake-bee-sub-1", nil)
	psMocks.ExpectDeleteTopic("workbench-notifications-fake-bee", nil)

	googleClients := mocks.NewClients(t)
	googleClients.On("PubSub", "bee-project").Return(psClient, nil)

	cleanup := NewCleanup(googleClients)
	require.NoError(t, cleanup.Cleanup(bee))
}

func Test_pubsubTopicIds(t *testing.T) {
	fixture := statefixtures.LoadFixture(statefixtures.Default, t)

	dynamic := fixture.Environment("fiab-funky-chipmunk")
	assert.ElementsMatch(t, []string{
		"leonardo-pubsub-fiab-funky-chipmunk",
		"rawls-async-import-topic-fiab-funky-chipmunk",
		"sam-group-sync-fiab-funky-chipmunk",
		"terra-fiab-funky-chipmunk-stairwaycluster-workqueue",
		"workbench-notifications-fiab-funky-chipmunk",
	}, pubsubTopicIds(dynamic))
}

func Test_projectIds(t *testing.T) {
	fixture := statefixtures.LoadFixture(statefixtures.Default, t)

	static := fixture.Environment("staging")
	assert.ElementsMatch(t, []string{"broad-dsde-staging", "terra-datarepo-staging"}, projectIds(static))

	dynamic := fixture.Environment("fiab-funky-chipmunk")
	assert.ElementsMatch(t, []string{"broad-dsde-qa"}, projectIds(dynamic))
}
