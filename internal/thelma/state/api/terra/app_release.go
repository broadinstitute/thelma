package terra

type AppRelease interface {
	AppVersion() string
	Environment() Environment
	// Subdomain returns the slug that this AppRelease uses inside its Environment. Defaults to chart Release.Name.
	// E.g. "leonardo"
	Subdomain() string
	// Protocol returns the transport protocol to connect to the AppRelease. Defaults to "https".
	Protocol() string
	// Port returns the transport protocol port to use to connect to the AppRelease. Defaults to '443'.
	Port() int
	// Host combines Subdomain with Environment.Destination.Name (if Environment.NamePrefixesDomain) and
	// Environment.BaseDomain to form the full hostname of the AppRelease.
	// E.g. "leonardo.my-environment.bee.envs-terra.bio"
	Host() string
	// URL combines Protocol with Host (no Port) to form the URL for the AppRelease.
	// E.g. "https://leonardo.my-environment.bee.envs-terra.bio"
	URL() string
	Release
}
