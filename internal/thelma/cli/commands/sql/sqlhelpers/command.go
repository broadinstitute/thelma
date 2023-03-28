package sqlhelpers

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/maps"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
)

// flagNames the names of all this command's CLI flags are kept in a struct so they can be easily referenced
var flagNames = struct {
	googleInstance  string
	googleProject   string
	chartRelease    string
	database        string
	permissionLevel string
}{
	googleInstance:  "google-instance",
	googleProject:   "google-project",
	chartRelease:    "chart-release",
	database:        "database",
	permissionLevel: "permission-level",
}

var permissionLevelNames = map[api.PermissionLevel][]string{
	api.ReadOnly:  {"read-only"},
	api.ReadWrite: {"read-write"},
	api.Admin:     {"admin"},
}

// command converts a child command that implements the sql.Command interface
// into a cli.ThelmaCommand
type command struct {
	child Command
	flags struct {
		googleInstance string
		googleProject  string
		chartRelease   string
	}
	opts api.ConnectionOptions
}

func (c *command) ConfigureCobra(cobraCommand *cobra.Command) {
	c.child.ConfigureCobra(cobraCommand)

	cobraCommand.Flags().StringVar(&c.flags.googleInstance, flagNames.googleInstance, "", `Name of Google CloudSQL instance`)
	cobraCommand.Flags().StringVar(&c.flags.googleProject, flagNames.googleProject, "", `Name of project containing Google CloudSQL instance`)
	cobraCommand.Flags().StringVar(&c.flags.chartRelease, flagNames.chartRelease, "", "Name of a chart release running in Kubernetes to connect to")
	cobraCommand.Flags().StringVarP(&c.opts.Database, flagNames.database, "d", "", "Database to connect to")

	permLevels := maps.ValuesFlattened(permissionLevelNames)
	cobraCommand.Flags().VarP(
		enumflag.New(&c.opts.PermissionLevel, flagNames.permissionLevel, permissionLevelNames, enumflag.EnumCaseInsensitive),
		flagNames.permissionLevel, "p",
		fmt.Sprintf("Permission level to connect with; one of: %s", utils.QuoteJoin(permLevels)),
	)
	if err := cobraCommand.RegisterFlagCompletionFunc(flagNames.permissionLevel, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return permLevels, cobra.ShellCompDirectiveDefault
	}); err != nil {
		panic(fmt.Errorf("failed to register cobra flag completion function for --%s: %v", flagNames.permissionLevel, err))
	}
}

func (c *command) PreRun(_ app.ThelmaApp, _ cli.RunContext) error {
	return nil
}

func (c *command) Run(app app.ThelmaApp, ctx cli.RunContext) error {
	var conn api.Connection
	conn.Options = c.opts

	flags := ctx.CobraCommand().Flags()

	if flags.Changed(flagNames.googleInstance) {
		if !flags.Changed(flagNames.googleProject) {
			return fmt.Errorf("--%s requires a google project be specified with --%s", flagNames.googleInstance, flagNames.googleProject)
		}
		conn.Provider = api.Google
		conn.GoogleInstance.Project = c.flags.googleProject
		conn.GoogleInstance.InstanceName = c.flags.googleInstance
		cluster, err := findProxyClusterForManuallySpecifiedDatabase(app, conn.GoogleInstance)
		if err != nil {
			return err
		}
		conn.Options.ProxyCluster = cluster
	} else if flags.Changed(flagNames.chartRelease) {
		state, err := app.State()
		if err != nil {
			return err
		}
		releases, err := state.Releases().Filter(filter.Releases().HasFullName(c.flags.chartRelease))
		if err != nil {
			return err
		}
		if len(releases) != 1 {
			return fmt.Errorf("found %d releases matching name %q, expected 1", len(releases), c.flags.chartRelease)
		}
		release := releases[0]
		conn.Provider = api.Kubernetes
		conn.KubernetesInstance.Release = release
		conn.Options.Release = release
		conn.Options.ProxyCluster = release.Cluster()
	} else {
		return fmt.Errorf("either --%s or --%s must be specified", flagNames.googleInstance, flagNames.chartRelease)
	}

	return c.child.Run(conn, app, ctx)
}

func (c *command) PostRun(app app.ThelmaApp, ctx cli.RunContext) error {
	return nil
}

func findProxyClusterForManuallySpecifiedDatabase(app app.ThelmaApp, cloudSQL api.GoogleInstance) (terra.Cluster, error) {
	state, err := app.State()
	if err != nil {
		return nil, fmt.Errorf("error loading state: %v", err)
	}

	clusters, err := state.Clusters().All()
	if err != nil {
		return nil, fmt.Errorf("error loading clusters from state: %v", err)
	}

	var inProject []terra.Cluster
	for _, c := range clusters {
		if c.Project() == cloudSQL.Project {
			inProject = append(inProject, c)
		}
	}

	if len(inProject) == 0 {
		return nil, fmt.Errorf("can't connect to CloudSQL instances in project %q (it has no Terra K8s cluster)", cloudSQL.Project)
	}

	// TODO - here we try to select a default cluster for projects that host multiple clusters.
	// Probably should be in sherlock or at least some formalized config struct somewhere in Thelma
	for _, c := range inProject {
		if cloudSQL.Project == "broad-dsde-qa" && c.Name() == "terra-qa-bees" {
			return c, nil
		}
		if cloudSQL.Project == "broad-dsde-dev" && c.Name() == "terra-dev" {
			return c, nil
		}
	}
	return inProject[0], nil
}
