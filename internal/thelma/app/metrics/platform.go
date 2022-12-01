package metrics

import (
	"fmt"
	"strings"
)

// Platform represents the kind of environment thelma is running in.
// For example: locally, on a laptop; in ArgoCD; in GitHub actions; etc.
type Platform int

const (
	Unknown Platform = iota
	Local
	ArgoCD
	GithubActions
	Jenkins
)

// String implement fmt.Stringer interface
func (p Platform) String() string {
	switch p {
	case Unknown:
		return "unknown"
	case Local:
		return "local"
	case ArgoCD:
		return "argocd"
	case GithubActions:
		return "gha"
	case Jenkins:
		return "jenkins"
	}
	panic(fmt.Errorf("unrecognized platform: %#v", p))
}

// UnmarshalText implement encoding.TextUnmarshaler interface so platform can be deserialized from config
func (p *Platform) UnmarshalText(text []byte) error {
	s := string(text)
	switch strings.ToLower(s) {
	case "unknown":
		*p = Unknown
	case "local":
		*p = Local
	case "argocd":
		*p = ArgoCD
	case "gha":
		*p = GithubActions
	case "jenkins":
		*p = Jenkins
	default:
		return fmt.Errorf("invalid platform: %q", s)
	}
	return nil
}
