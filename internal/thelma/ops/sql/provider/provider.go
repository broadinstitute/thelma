package provider

import (
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/dbms"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun"
)

// Provider abstracts provider-specific features for Google CloudSQL, Azure, or K8s that are related to
// initializing and connecting to a database instance.
// (Note that a Provider is coupled to a specific api.Connection).
type Provider interface {
	// ClientSettings returns client settings that should be used to connect within the pod to the target instance
	ClientSettings(...ConnectionOverride) (dbms.ClientSettings, error)
	// DetectDBMS detects the instance's DBMS (MySQL or Postgres)
	DetectDBMS() (api.DBMS, error)
	// Initialized returns true if the database instance has been initialized, false otherwise
	Initialized() (bool, error)
	// Initialize performs any necessary initialization to set up the instance for future connections
	Initialize() error
	// PodSpec returns information about Kubernetes resources that should be created to connect to the database instance
	PodSpec(...ConnectionOverride) (podrun.ProviderSpec, error)
}

// ConnectionOverride generate client settings, but overriding parameters in the connection object
type ConnectionOverride func(options *api.ConnectionOptions)
