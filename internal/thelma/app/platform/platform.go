package platform

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
	"os/user"
	"runtime"
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

// local username of whoever invoked the legacy thelma docker wrapper
const wrapperUserEnvVar = "LEGACY_WRAPPER_USER"

// local username that processes run as in ArgoCD containers
const argocdUser = "argocd"

// name of an environment variable set in ArgoCD
// https://argo-cd.readthedocs.io/en/stable/user-guide/build-environment/
const argocdEnvVar = "ARGOCD_APP_NAME"

// local username that Jenkins nodes run as
const jenkinsUser = "jenkins"

// name of an environment variable set in GitHub actions
// https://docs.github.com/en/actions/learn-github-actions/environment-variables
const githubWorkflowEnvVar = "GITHUB_WORKFLOW"

// Lookup best-effort attempt to guess platform based on the environment thelma is running in
func Lookup() Platform {
	// Usage on a Mac is always local
	//goland:noinspection ALL
	if runtime.GOOS == "darwin" {
		return Local
	}

	// ArgoCD sidecar plugins can run as user 999, so calling user.Current() will fail
	// (Presumably due to lack of username, though the stated error complains about
	// `user: Current requires cgo or $USER set in environment`)
	// https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/#register-the-plugin-sidecar:~:text=Make%20sure%20that%20sidecar%20container%20is%20running%20as%20user%20999
	//
	// We can still identify ArgoCD by its build environment variables, and exit out
	// before calling user.Current()
	// https://argo-cd.readthedocs.io/en/stable/user-guide/build-environment/
	if _, present := os.LookupEnv(argocdEnvVar); present {
		return ArgoCD
	}

	u, err := user.Current()
	if err != nil {
		log.Warn().Err(err).Msgf("failed to identify process owner")
		return Unknown
	}

	// ArgoCD configmap plugins run as the ArgoCD user
	// https://github.com/argoproj/argo-cd/blob/master/Dockerfile#L76
	if u.Username == argocdUser {
		return ArgoCD
	}

	// GitHub sets the following environment variables when running containers in Actions
	// https://docs.github.com/en/actions/learn-github-actions/environment-variables
	if os.Getenv(githubWorkflowEnvVar) != "" {
		return GithubActions
	}

	// Jenkins runs thelma using the legacy docker wrapper in terra-helmfile
	if os.Getenv(wrapperUserEnvVar) == jenkinsUser {
		return Jenkins
	}

	return Unknown
}

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
	panic(errors.Errorf("unrecognized platform: %#v", p))
}

// Link returns a link to the CI/CD logs for this Thelma run, if applicable
func (p Platform) Link() string {
	if p == GithubActions {
		// ref: https://docs.github.com/en/actions/learn-github-actions/variables
		// $GITHUB_SERVER_URL/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID
		return fmt.Sprintf("%s/%s/actions/runs/%s", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_ID"))
	}
	// TODO - return Jenkins and ArgoCD links at some point?
	return ""
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
		return errors.Errorf("invalid platform: %q", s)
	}
	return nil
}
