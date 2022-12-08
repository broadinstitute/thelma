package testing

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/testing/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PubSubMocks struct {
	projectId        string
	publisherServer  *mocks.PublisherServer
	subscriberServer *mocks.SubscriberServer
}

func (p *PubSubMocks) PublisherServer() *mocks.PublisherServer {
	return p.publisherServer
}

func (p *PubSubMocks) SubscriberServer() *mocks.SubscriberServer {
	return p.subscriberServer
}

func (p *PubSubMocks) ExpectGetTopic(topicId string, resp *pubsubpb.Topic, err error) {
	p.publisherServer.On("GetTopic", mock.Anything, &pubsubpb.GetTopicRequest{
		Topic: p.topicName(topicId),
	}).Return(resp, err)
}

func (p *PubSubMocks) ExpectListTopicSubscriptions(topicId string, subscriptionIDsToReturn []string, err error) {
	var subscriptions []string
	for _, id := range subscriptionIDsToReturn {
		subscriptions = append(subscriptions, p.subscriptionName(id))
	}

	p.publisherServer.On("ListTopicSubscriptions", mock.Anything, &pubsubpb.ListTopicSubscriptionsRequest{
		Topic: p.topicName(topicId),
	}).Return(&pubsubpb.ListTopicSubscriptionsResponse{
		Subscriptions: subscriptions,
	}, err)
}

func (p *PubSubMocks) ExpectDeleteSubscription(subscriptionId string, err error) {
	p.subscriberServer.On("DeleteSubscription", mock.Anything, &pubsubpb.DeleteSubscriptionRequest{
		Subscription: p.subscriptionName(subscriptionId),
	}).Return(&emptypb.Empty{}, err)
}

func (p *PubSubMocks) ExpectDeleteTopic(topicId string, err error) {
	p.publisherServer.On("DeleteTopic", mock.Anything, &pubsubpb.DeleteTopicRequest{
		Topic: p.topicName(topicId),
	}).Return(&emptypb.Empty{}, err)
}

func (p *PubSubMocks) topicName(topicId string) string {
	return fmt.Sprintf("projects/%s/topics/%s", p.projectId, topicId)
}

func (p *PubSubMocks) subscriptionName(subscriptionId string) string {
	return fmt.Sprintf("projects/%s/subcriptions/%s", p.projectId, subscriptionId)
}

// NewMockPubSubServerAndClient creates a connected mock cluster manager server and client
func NewMockPubSubServerAndClient(t *testing.T, projectId string) (*PubSubMocks, *pubsub.Client) {
	// Note - we intentionally do NOT pass in the test object to the generated mock, because it
	// will end up calling t.FailNow() in the grpc server's goroutine instead of main,
	// causing tests to hang in some situations. (It's a no-no to call t.FailNow() outside of the test's main goroutine)
	mockPublisherServer := &mocks.PublisherServer{}
	mockSubscriberServer := &mocks.SubscriberServer{}

	t.Cleanup(func() {
		mockPublisherServer.AssertExpectations(t)
		mockSubscriberServer.AssertExpectations(t)
	})

	server := newFakeGRPCServer(t, func(s *grpc.Server) {
		// Configure the GRPC server to serve responses from our mock server
		pubsubpb.RegisterPublisherServer(s, mockPublisherServer)
		pubsubpb.RegisterSubscriberServer(s, mockSubscriberServer)
	})

	client, err := pubsub.NewClient(context.TODO(), projectId, server.clientOptions...)
	require.NoError(t, err)

	return &PubSubMocks{
		projectId:        projectId,
		publisherServer:  mockPublisherServer,
		subscriberServer: mockSubscriberServer,
	}, client
}
