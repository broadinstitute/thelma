package selector

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
)

// -r / --release flag
func newReleasesFlag() *enumFlag {
	return &enumFlag{
		flagName:      flagNames.release,
		shortHand:     "r",
		defaultValues: []string{},
		usageMessage:  "Run for specific release(s) (set to ALL to include all releases)",

		preProcessHook: func(flagValues []string, args []string, changed bool) ([]string, error) {
			// UX: make it possible for users to specify a release as the first positional arg instead of as a flag
			if changed {
				if len(args) > 0 {
					return nil, fmt.Errorf("releases can either be specified with the --%s flag or via positional argument, not both", ReleasesFlagName)
				}
				return flagValues, nil
			} else if len(args) > 0 {
				return []string{args[0]}, nil
			} else {
				// We have a lot of releases, and most developers want to render for a specific service,
				// so force users to supply --releases=ALL if they _really_ want to render for all releases and not a specific release
				return nil, fmt.Errorf(`please specify at least one release with --%s <RELEASE>, or select all releases with --%s %s`, ReleasesFlagName, ReleasesFlagName, allSelector)
			}
		},

		validValues: func(state terra.State) (set.StringSet, error) {
			releases, err := state.Releases().All()
			if err != nil {
				return nil, err
			}
			return releaseNames(releases), nil
		},

		buildFilter: func(f *filterBuilder, uniqueValues []string) {
			f.addReleaseFilter(filter.Releases().HasName(uniqueValues...))
		},
	}
}

// -c / --cluster flag
func newClustersFlag() *enumFlag {
	return &enumFlag{
		flagName:      flagNames.cluster,
		shortHand:     "c",
		defaultValues: []string{allSelector},
		usageMessage:  "Run for specific Terra cluster(s)",

		validValues: func(state terra.State) (set.StringSet, error) {
			clusters, err := state.Clusters().All()
			if err != nil {
				return nil, err
			}
			return clusterNames(clusters), nil
		},

		buildFilter: func(f *filterBuilder, uniqueValues []string) {
			f.addDestinationInclude(filter.Destinations().IsCluster().And(filter.Destinations().HasName(uniqueValues...)))
		},
	}
}

// -e / --environment flag
func newEnvironmentsFlag() *enumFlag {
	return &enumFlag{
		flagName:      flagNames.environment,
		shortHand:     "e",
		defaultValues: []string{allSelector},
		usageMessage:  "Run for specific Terra environment(s)",

		validValues: func(state terra.State) (set.StringSet, error) {
			environments, err := state.Environments().All()
			if err != nil {
				return nil, err
			}
			return environmentNames(environments), nil
		},

		buildFilter: func(f *filterBuilder, uniqueValues []string) {
			f.addDestinationInclude(filter.Destinations().IsEnvironment().And(filter.Destinations().HasName(uniqueValues...)))
		},
	}
}

// --destination-type flag
func newDestinationTypesFlag() *enumFlag {
	return &enumFlag{
		flagName:      flagNames.destinationType,
		defaultValues: []string{allSelector},
		usageMessage:  `Run for a specific destination type (eg. "environment". "cluster"`,

		validValues: func(_ terra.State) (set.StringSet, error) {
			return set.NewStringSet(terra.DestinationTypeNames()...), nil
		},

		buildFilter: func(f *filterBuilder, uniqueValues []string) {
			f.addDestinationFilter(filter.Destinations().OfTypeName(uniqueValues...))
		},
	}
}

// --destination-base flag
func newDestinationBasesFlag() *enumFlag {
	return &enumFlag{
		flagName:      flagNames.destinationBase,
		defaultValues: []string{allSelector},
		usageMessage:  `Run for a specific environment or cluster base (eg. \"live\", \"bee\")`,

		validValues: func(state terra.State) (set.StringSet, error) {
			destinations, err := state.Destinations().All()
			if err != nil {
				return nil, err
			}

			s := set.NewStringSet()
			for _, d := range destinations {
				s.Add(d.Base())
			}
			return s, nil
		},

		buildFilter: func(f *filterBuilder, uniqueValues []string) {
			f.addDestinationFilter(filter.Destinations().HasBase(uniqueValues...))
		},
	}
}

// --environment-templates flag
func newEnvironmentTemplatesFlag() *enumFlag {
	return &enumFlag{
		flagName:      flagNames.environmentTemplate,
		defaultValues: []string{allSelector},
		usageMessage:  `Run for dynamic environments with a specific template (eg. "swatomation")`,

		validValues: func(state terra.State) (set.StringSet, error) {
			envs, err := state.Environments().Filter(filter.Environments().HasLifecycle(terra.Template))
			if err != nil {
				return nil, err
			}
			return environmentNames(envs), nil
		},

		buildFilter: func(f *filterBuilder, values []string) {
			f.addEnvironmentFilter(filter.Environments().HasTemplateName(values...))
		},
	}
}

// --environment-lifecycles flag
func newEnvironmentLifecyclesFlag() *enumFlag {
	return &enumFlag{
		flagName:      flagNames.environmentLifecycle,
		defaultValues: []string{terra.Static.String(), terra.Template.String()},
		usageMessage:  `Run for environments with a specific lifecycle (eg. "static", "template", "dynamic")`,

		validValues: func(_ terra.State) (set.StringSet, error) {
			return set.NewStringSet(terra.LifecycleNames()...), nil
		},

		buildFilter: func(f *filterBuilder, uniqueValues []string) {
			f.addEnvironmentFilter(filter.Environments().HasLifecycleName(uniqueValues...))
		},
	}
}

// TODO refactor these repetitive methods when generics are available

func releaseNames(releases []terra.Release) set.StringSet {
	names := set.NewStringSet()
	for _, r := range releases {
		names.Add(r.Name())
	}
	return names
}

func environmentNames(envs []terra.Environment) set.StringSet {
	names := set.NewStringSet()
	for _, env := range envs {
		names.Add(env.Name())
	}
	return names
}

func clusterNames(clusters []terra.Cluster) set.StringSet {
	names := set.NewStringSet()
	for _, cluster := range clusters {
		names.Add(cluster.Name())
	}
	return names
}
