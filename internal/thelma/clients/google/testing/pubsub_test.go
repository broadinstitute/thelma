package testing

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"
)

func Test_PubSub(t *testing.T) {
	mocks, client := NewMockPubSubServerAndClient(t, "my-fake-project")

	mocks.ExpectGetTopic("my-topic-id", &pubsubpb.Topic{
		Name: "my-topic-id",
	}, nil)

	mocks.ExpectGetTopic("does-not-exist", nil, NotFoundError())

	mocks.ExpectListTopicSubscriptions("my-topic-id", []string{
		"sub-1",
		"sub-2",
	}, nil)

	mocks.ExpectDeleteSubscription("sub-1", nil)
	mocks.ExpectDeleteSubscription("sub-2", nil)
	mocks.ExpectDeleteTopic("my-topic-id", nil)

	topic := client.Topic("my-topic-id")
	exists, err := topic.Exists(context.TODO())
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "my-topic-id", topic.ID())

	missing := client.Topic("does-not-exist")
	exists, err = missing.Exists(context.TODO())
	require.NoError(t, err)
	assert.False(t, exists)

	var subscriptions []*pubsub.Subscription
	it := topic.Subscriptions(context.TODO())
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			require.NoError(t, err)
		}
		subscriptions = append(subscriptions, sub)
	}

	require.Equal(t, 2, len(subscriptions))

	sub1 := subscriptions[0]
	sub2 := subscriptions[1]

	assert.Equal(t, "sub-1", sub1.ID())
	assert.Equal(t, "sub-2", sub2.ID())

	require.NoError(t, sub1.Delete(context.TODO()))
	require.NoError(t, sub2.Delete(context.TODO()))
	require.NoError(t, topic.Delete(context.TODO()))
}
