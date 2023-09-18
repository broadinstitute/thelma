// Package filetrigger maps a list of updated files in the terra-helmfile repo to a list of charts that need to be published.
package filetrigger

import "github.com/broadinstitute/thelma/internal/thelma/state/api/terra"

const FlagName = "file-trigger"

func ReleaseFilter(filetriggers ...string) terra.ReleaseFilter {
	panic("TODO")
}

func ChartList(filetrigger string, state terra.State) ([]string, error) {
	panic("TODO")
}
