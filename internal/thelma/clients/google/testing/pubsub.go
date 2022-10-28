package testing

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/clients/google/testing/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	pubsubpb "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"testing"
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
	mockPublisherServer := mocks.NewPublisherServer(t)
	mockSubscriberServer := mocks.NewSubscriberServer(t)

	// Create a GRPC server
	// reference: https://github.com/googleapis/google-cloud-go/blob/main/testing.md#testing-grpc-services-using-fakes
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	gsrv := grpc.NewServer()
	t.Cleanup(func() {
		gsrv.Stop()
	})

	// Configure the GRPC server to serve responses from our mock server
	pubsubpb.RegisterPublisherServer(gsrv, mockPublisherServer)
	pubsubpb.RegisterSubscriberServer(gsrv, mockSubscriberServer)
	//	pubsubpb.RegisterSubscriberServer()
	go func() {
		if err := gsrv.Serve(listener); err != nil {
			panic(err)
		}
	}()

	// Create client that is configured to talk to the fake GRPC server
	fakeServerAddr := listener.Addr().String()
	client, err := pubsub.NewClient(context.TODO(),
		projectId,
		option.WithEndpoint(fakeServerAddr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	require.NoError(t, err)

	return &PubSubMocks{
		projectId:        projectId,
		publisherServer:  mockPublisherServer,
		subscriberServer: mockSubscriberServer,
	}, client
}
