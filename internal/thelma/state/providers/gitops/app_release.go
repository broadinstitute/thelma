package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"strings"
)

// implements the terra.AppRelease interface
type appRelease struct {
	appVersion string
	subdomain  string
	protocol   string
	port       int
	release
}

func (r *appRelease) AppVersion() string {
	return r.appVersion
}

func (r *appRelease) Environment() terra.Environment {
	return r.destination.(terra.Environment)
}

func (r *appRelease) Subdomain() string {
	if r.subdomain == "" {
		return r.chartName
	} else {
		return r.subdomain
	}
}

func (r *appRelease) Protocol() string {
	if r.protocol == "" {
		return "https"
	} else {
		return r.protocol
	}
}

func (r *appRelease) Port() int {
	if r.port == 0 {
		return 443
	} else {
		return r.port
	}
}

func (r *appRelease) Host() string {
	var parts []string
	parts = append(parts, r.Subdomain())
	if r.Environment().NamePrefixesDomain() {
		parts = append(parts, r.Environment().Name())
	}
	if r.Environment().BaseDomain() != "" {
		parts = append(parts, r.Environment().BaseDomain())
	}
	return strings.Join(parts, ".")
}

func (r *appRelease) URL() string {
	return fmt.Sprintf("%s://%s", r.Protocol(), r.Host())
}
