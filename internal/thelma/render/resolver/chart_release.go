package resolver

// ChartRelease the set of attributes needed to find and download a Helm chart
type ChartRelease struct {
	Name    string
	Repo    string
	Version string
}
