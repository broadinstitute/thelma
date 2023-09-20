package api

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"strings"
)

// Connection parameters for a thelma sql connection
type Connection struct {
	// Provider where the database is running
	Provider Provider
	// GoogleInstance CloudSQL connection parameters
	GoogleInstance GoogleInstance
	// AzureInstance Azure connection parameters
	AzureInstance AzureInstance
	// KubernetesInstance Kubernetes connection parameters
	KubernetesInstance KubernetesInstance
	// Options cross-platform connection options
	Options ConnectionOptions
}

func (c Connection) Instance() Instance {
	switch c.Provider {
	case Google:
		return c.GoogleInstance
	case Azure:
		return c.AzureInstance
	case Kubernetes:
		return c.KubernetesInstance
	default:
		panic(errors.Errorf("unknown platform: %#v", c.Provider))
	}
}

type Instance interface {
	// Name returns a descriptive name for the instance (eg. CloudSQL instance name)
	Name() string
	// IsProd returns true if this instance is in production
	IsProd() bool
	// IsShared returns true if this is a shared instance
	IsShared() bool
}

// GoogleInstance connection parameters for a Google CloudSQL instance
type GoogleInstance struct {
	InstanceName string // required; CloudSQL instance name
	Project      string // required; CloudSQL google project name
}

func (g GoogleInstance) Name() string {
	return g.InstanceName
}

func (g GoogleInstance) IsProd() bool {
	// TODO make more robust (model in Sherlock?)
	return strings.HasSuffix(g.Project, "prod") || strings.HasSuffix(g.Project, "production")
}

func (g GoogleInstance) IsShared() bool {
	// TODO make more robust (model in Sherlock?)
	return strings.HasPrefix(g.Project, "broad-dsde")
}

// AzureInstance connection parameters for an Azure db
type AzureInstance struct {
	// TODO
}

func (a AzureInstance) Name() string {
	//TODO implement me
	panic("implement me")
}

func (a AzureInstance) IsProd() bool {
	panic("TODO")
}

func (a AzureInstance) IsShared() bool {
	panic("TODO")
}

// KubernetesInstance connection parameters for a K8s-hosted db instance
type KubernetesInstance struct {
	Release terra.Release
}

func (k KubernetesInstance) Name() string {
	return k.Release.FullName()
}

func (k KubernetesInstance) IsProd() bool {
	return k.Release.Cluster().RequireSuitable()
}

func (k KubernetesInstance) IsShared() bool {
	if !k.Release.IsAppRelease() {
		// cluster releases are "shared" in that they aren't part of a BEE
		return true
	}
	// TODO not perfect (microsoft BEEs are shared)
	return k.Release.(terra.AppRelease).Environment().Lifecycle() != terra.Dynamic
}

type ConnectionOptions struct {
	// Database (within the instance) to connect to
	Database string
	// PrivilegeLevel permission level to use for connection
	PrivilegeLevel PrivilegeLevel
	// ProxyCluster terra Kubernetes cluster to connect through
	ProxyCluster terra.Cluster
	// Release (nil if the target database instance has no association with a Terra release)
	Release terra.Release
	// Shell for `thelma sql connect`, optionally launch an interactive Bash shell instead of the default psql/mysql command
	Shell bool
}

type PrivilegeLevel int64

const (
	ReadOnly PrivilegeLevel = iota
	ReadWrite
	Admin
)
