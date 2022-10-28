package generators

// This file uses mockery to generate mocks for interfaces in Google cloud libraries as described here:
// https://github.com/vektra/mockery/issues/243#issuecomment-1143371435

//go:generate mockery --srcpkg=google.golang.org/genproto/googleapis/container/v1 --name=ClusterManagerServer --output=../mocks --filename=cluster_manager_server.go
//go:generate mockery --srcpkg=google.golang.org/genproto/googleapis/pubsub/v1 --name=PublisherServer --output=../mocks --filename=publisher_server.go
//go:generate mockery --srcpkg=google.golang.org/genproto/googleapis/pubsub/v1 --name=SubscriberServer --output=../mocks --filename=subscriber_server.go
