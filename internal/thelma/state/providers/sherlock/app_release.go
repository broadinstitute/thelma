package sherlock

import (
	"fmt"
	"strings"

	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
)

type appRelease struct {
	release
}

func (r *appRelease) Environment() terra.Environment {
	return r.destination.(terra.Environment)
}

func (r *appRelease) Subdomain() string {
	if r.subdomain == "" {
		return r.chartName
	}
	return r.subdomain
}

func (r *appRelease) Protocol() string {
	if r.protocol == "" {
		return "https"
	}
	return r.protocol
}

func (r *appRelease) Port() int {
	if r.port == 0 {
		return 443
	}
	return r.port
}

func (r *appRelease) Host() string {
	var components []string
	components = append(components, r.Subdomain())
	if r.Environment().NamePrefixesDomain() {
		components = append(components, r.Environment().Name())
	}

	if r.Environment().BaseDomain() != "" {
		components = append(components, r.Environment().BaseDomain())
	}
	return strings.Join(components, ".")
}

func (r *appRelease) URL() string {
	return fmt.Sprintf("%s://%s", r.Protocol(), r.Host())
}
