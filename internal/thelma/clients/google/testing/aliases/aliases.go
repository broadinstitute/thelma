package aliases

import (
	"cloud.google.com/go/container/apiv1/containerpb"
	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
)

// We create aliases in this file for the types in the Google client library that we want to mock,
// so that we can generate mocks for them with mockery

type ClusterManagerServer interface {
	containerpb.ClusterManagerServer
}

type PublisherServer interface {
	pubsubpb.PublisherServer
}

type SubscriberServer interface {
	pubsubpb.SubscriberServer
}
