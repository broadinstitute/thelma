package health

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

type Status struct {
}

type Reporter interface {
	GetStatus(releases []terra.Release) map[terra.Release]Status
}

//
// so, let's start by grabbig a list of resources from the ArgoCD application and events from the namespace with kubectl.
//
// probably what we can do is get all events for a group of releases? Whatever, it doesn't need to be super clean.
//
