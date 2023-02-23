package interfaces

import (
	"cloud.google.com/go/container/apiv1/containerpb"
	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
)

// We create aliases in this files for the types in the Google client library that we want to mock,
// so that we can generate mocks with mockery

type ClusterManagerServer interface {
	containerpb.ClusterManagerServer
}

type PublisherServer interface {
	pubsubpb.PublisherServer
}

type SubscriberServer interface {
	pubsubpb.SubscriberServer
}
