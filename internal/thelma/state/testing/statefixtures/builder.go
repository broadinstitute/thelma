package statefixtures

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"strings"
	"time"
)

type builder struct {
	data              *FixtureData
	environmentSet    map[string]*statemocks.Environment
	clusterSet        map[string]*statemocks.Cluster
	appReleaseSet     map[string]*statemocks.AppRelease
	clusterReleaseSet map[string]*statemocks.ClusterRelease
	clusters          *statemocks.Clusters
	environments      *StubEnvironments
	releases          *StubReleases
	state             *statemocks.State
	stateLoader       *statemocks.StateLoader
}

func newBuilder(data *FixtureData) *builder {
	return &builder{
		data:              data,
		environmentSet:    make(map[string]*statemocks.Environment),
		clusterSet:        make(map[string]*statemocks.Cluster),
		appReleaseSet:     make(map[string]*statemocks.AppRelease),
		clusterReleaseSet: make(map[string]*statemocks.ClusterRelease),
		clusters:          new(statemocks.Clusters),
		environments:      &StubEnvironments{Environments: new(statemocks.Environments)},
		releases:          &StubReleases{Releases: new(statemocks.Releases)},
		state:             new(statemocks.State),
		stateLoader:       new(statemocks.StateLoader),
	}
}

func (b *builder) buildMocks() *Mocks {
	// set up individual clusters/releases/
	b.populateClusterSet()
	b.populateEnvironmentSet()
	b.populateReleaseSets()
	b.addAppReleasesToEnvMocks()
	b.addClusterReleasesToClusterMocks()

	// set up clients (state.Clusters(), state.Releases(), etc)
	b.setClustersMocks()
	b.setEnvironmentsMocks()
	b.setReleasesMocks()

	// set up root objects
	b.setStateMocks()
	b.setStateLoaderMocks()

	return &Mocks{
		Clusters:     b.clusters,
		Environments: b.environments,
		Releases:     b.releases,
		State:        b.state,
		StateLoader:  b.stateLoader,
		Items: struct {
			Clusters        map[string]*statemocks.Cluster
			Environments    map[string]*statemocks.Environment
			AppReleases     map[string]*statemocks.AppRelease
			ClusterReleases map[string]*statemocks.ClusterRelease
		}{
			Clusters:        b.clusterSet,
			Environments:    b.environmentSet,
			AppReleases:     b.appReleaseSet,
			ClusterReleases: b.clusterReleaseSet,
		},
	}
}

func (b *builder) populateClusterSet() {
	for _, c := range b.data.Clusters {
		cluster := &statemocks.Cluster{}
		cluster.EXPECT().Name().Return(c.Name)
		cluster.EXPECT().Base().Return(c.Base)
		cluster.EXPECT().Location().Return(c.Location)
		cluster.EXPECT().Project().Return(c.Project)
		cluster.EXPECT().Address().Return(c.Address)
		cluster.EXPECT().RequireSuitable().Return(c.RequireSuitable)
		cluster.EXPECT().Type().Return(terra.ClusterDestination)
		cluster.EXPECT().IsEnvironment().Return(false)
		cluster.EXPECT().ReleaseType().Return(terra.ClusterReleaseType)
		cluster.EXPECT().ArtifactBucket().Return(fmt.Sprintf("thelma-artifacts-%s", c.Name))

		tokens := strings.Split(c.Project, "-")
		suffix := tokens[len(tokens)-1:][0]
		cluster.EXPECT().ProjectSuffix().Return(suffix)

		cluster.EXPECT().TerraHelmfileRef().Return(c.TerraHelmfileRef)

		b.clusterSet[c.Name] = cluster
	}
}

func (b *builder) populateEnvironmentSet() {
	for _, e := range b.data.Environments {
		env := &statemocks.Environment{}
		env.EXPECT().Name().Return(e.Name)
		env.EXPECT().Base().Return(e.Base)
		env.EXPECT().DefaultCluster().Return(b.clusterSet[e.DefaultCluster])
		env.EXPECT().Lifecycle().Return(e.Lifecycle)
		env.EXPECT().Template().Return(e.Template)
		env.EXPECT().UniqueResourcePrefix().Return(e.UniqueResourcePrefix)
		env.EXPECT().RequireSuitable().Return(e.RequireSuitable)
		env.EXPECT().Type().Return(terra.EnvironmentDestination)
		env.EXPECT().IsEnvironment().Return(true)
		env.EXPECT().ReleaseType().Return(terra.AppReleaseType)
		env.EXPECT().TerraHelmfileRef().Return(e.TerraHelmfileRef)
		env.EXPECT().BaseDomain().Return("")
		env.EXPECT().Namespace().Return("terra-" + e.Name)
		env.EXPECT().NamePrefixesDomain().Return(true)
		env.EXPECT().PreventDeletion().Return(false)
		env.EXPECT().Owner().Return(e.Owner)

		autodelete := new(statemocks.AutoDelete)
		autodelete.EXPECT().Enabled().Return(false)
		autodelete.EXPECT().After().Return(time.Time{})
		env.EXPECT().AutoDelete().Return(autodelete)
		b.environmentSet[e.Name] = env
	}
}

func (b *builder) populateReleaseSets() {
	for _, r := range b.data.Releases {
		if r.Environment != "" {
			b.addAppReleaseToSet(r)
		} else {
			b.addClusterReleaseToSet(r)
		}
	}
}

func (b *builder) addAppReleaseToSet(r Release) {
	release := &statemocks.AppRelease{}
	release.EXPECT().AppVersion().Return(r.AppVersion)
	release.EXPECT().ChartName().Return(r.Chart)
	release.EXPECT().ChartVersion().Return(r.ChartVersion)
	release.EXPECT().Cluster().Return(b.clusterSet[r.Cluster])
	release.EXPECT().ClusterAddress().Return(b.clusterSet[r.Cluster].Address())
	release.EXPECT().ClusterName().Return(b.clusterSet[r.Cluster].Name())
	release.EXPECT().Destination().Return(b.environmentSet[r.Environment])
	release.EXPECT().Environment().Return(b.environmentSet[r.Environment])
	release.EXPECT().FullName().Return(r.FullName)
	release.EXPECT().HelmfileOverlays().Return(nil)
	release.EXPECT().IsAppRelease().Return(true)
	release.EXPECT().IsClusterRelease().Return(false)
	release.EXPECT().Name().Return(r.name())
	release.EXPECT().Namespace().Return(r.Namespace)
	release.EXPECT().Port().Return(r.Port)
	release.EXPECT().Protocol().Return(r.Protocol)
	release.EXPECT().Repo().Return(r.Repo)
	release.EXPECT().Subdomain().Return(r.Subdomain)
	release.EXPECT().TerraHelmfileRef().Return(r.TerraHelmfileRef)
	release.EXPECT().Type().Return(terra.AppReleaseType)
	b.appReleaseSet[r.key()] = release
}

func (b *builder) addClusterReleaseToSet(r Release) {
	release := &statemocks.ClusterRelease{}
	release.EXPECT().AppVersion().Return(r.AppVersion)
	release.EXPECT().ChartName().Return(r.Chart)
	release.EXPECT().ChartVersion().Return(r.ChartVersion)
	release.EXPECT().Cluster().Return(b.clusterSet[r.Cluster])
	release.EXPECT().ClusterAddress().Return(b.clusterSet[r.Cluster].Address())
	release.EXPECT().ClusterName().Return(b.clusterSet[r.Cluster].Name())
	release.EXPECT().Destination().Return(b.clusterSet[r.Cluster])
	release.EXPECT().FullName().Return(r.FullName)
	release.EXPECT().HelmfileOverlays().Return(nil)
	release.EXPECT().IsAppRelease().Return(false)
	release.EXPECT().IsClusterRelease().Return(true)
	release.EXPECT().Name().Return(r.name())
	release.EXPECT().Namespace().Return(r.Namespace)
	release.EXPECT().Repo().Return(r.Repo)
	release.EXPECT().TerraHelmfileRef().Return(r.TerraHelmfileRef)
	release.EXPECT().Type().Return(terra.ClusterReleaseType)
	b.clusterReleaseSet[r.key()] = release
}

func (b *builder) addClusterReleasesToClusterMocks() {
	for _, c := range b.data.Clusters {
		cluster := b.clusterSet[c.Name]

		var releases []terra.Release
		for _, r := range b.data.Releases {
			if r.Environment != "" {
				continue
			}
			if r.Cluster != c.Name {
				continue
			}
			releases = append(releases, b.clusterReleaseSet[r.key()])
		}

		cluster.EXPECT().Releases().Return(releases)
	}
}

func (b *builder) addAppReleasesToEnvMocks() {
	for _, e := range b.data.Environments {
		env := b.environmentSet[e.Name]

		var releases []terra.Release
		for _, r := range b.data.Releases {
			if r.Environment == "" {
				continue
			}
			if r.Environment != e.Name {
				continue
			}
			releases = append(releases, b.appReleaseSet[r.key()])
		}

		env.EXPECT().Releases().Return(releases)
	}
}

func (b *builder) setClustersMocks() {
	var allClusters []terra.Cluster

	for clusterName, cluster := range b.clusterSet {
		b.clusters.EXPECT().Get(clusterName).Return(cluster, nil)
		b.clusters.EXPECT().Exists(clusterName).Return(true, nil)
		allClusters = append(allClusters, cluster)
	}

	b.clusters.EXPECT().All().Return(allClusters, nil)
}

func (b *builder) setEnvironmentsMocks() {
	var allEnvironments []terra.Environment

	for envName, env := range b.environmentSet {
		b.environments.EXPECT().Get(envName).Return(env, nil)
		b.environments.EXPECT().Exists(envName).Return(true, nil)
		allEnvironments = append(allEnvironments, env)
	}

	b.environments.EXPECT().All().Return(allEnvironments, nil)
}

func (b *builder) setReleasesMocks() {
	var allReleases []terra.Release

	for _, release := range b.appReleaseSet {
		allReleases = append(allReleases, release)
	}
	for _, release := range b.clusterReleaseSet {
		allReleases = append(allReleases, release)
	}

	b.releases.EXPECT().All().Return(allReleases, nil)
}

func (b *builder) setStateMocks() {
	b.state.EXPECT().Environments().Return(b.environments)
	b.state.EXPECT().Clusters().Return(b.clusters)
	b.state.EXPECT().Releases().Return(b.releases)
}

func (b *builder) setStateLoaderMocks() {
	b.stateLoader.EXPECT().Load().Return(b.state, nil)
	b.stateLoader.EXPECT().Reload().Return(b.state, nil)
}
