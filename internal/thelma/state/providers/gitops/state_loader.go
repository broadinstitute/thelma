package gitops

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/deepmerge"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/serializers"
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statebucket"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type LoadOptions struct {
	StateBucket statebucket.StateBucket
}

type LoadOption func(LoadOptions) LoadOptions

// Default settings file name for both types of destinations
const defaultsFileName = "defaults.yaml"
const defaultChartRepo = "terra-helm"
const defaultEnvironmentLifecycle = terra.Static
const yamlSuffix = ".yaml"

// envConfigDir is the subdirectory in terra-helmfile to search for environment config files
const envConfigDir = "environments"

// clusterConfigDir is the subdirectory in terra-helmfile to search for cluster config files
const clusterConfigDir = "clusters"

// NewStateLoader returns a new StateLoader w/ given settings
func NewStateLoader(thelmaHome string, shellRunner shell.Runner, options ...LoadOption) (terra.StateLoader, error) {
	opts := LoadOptions{}
	for _, option := range options {
		opts = option(opts)
	}

	loader := &stateLoader{
		thelmaHome:  thelmaHome,
		shellRunner: shellRunner,
		statebucket: opts.StateBucket,
	}

	if loader.statebucket == nil {
		sb, err := statebucket.New()
		if err != nil {
			return nil, err
		}
		loader.statebucket = sb
	}

	return loader, nil
}

// implements terra.StateLoader interface
type stateLoader struct {
	statebucket statebucket.StateBucket
	thelmaHome  string
	shellRunner shell.Runner
}

func (s *stateLoader) Load() (terra.State, error) {
	_versions, err := NewVersions(s.thelmaHome, s.shellRunner)
	if err != nil {
		return nil, err
	}

	_clusters, err := loadClusters(s.thelmaHome, _versions)
	if err != nil {
		return nil, err
	}

	_environments, err := loadEnvironments(s.thelmaHome, _versions, _clusters, s.statebucket)
	if err != nil {
		return nil, err
	}

	return &state{
		statebucket:  s.statebucket,
		versions:     _versions,
		clusters:     _clusters,
		environments: _environments,
	}, nil
}

func loadEnvironments(configRepoPath string, versions Versions, clusters map[string]terra.Cluster, sb statebucket.StateBucket) (map[string]terra.Environment, error) {
	yamlEnvs, err := loadYamlEnvironments(configRepoPath, versions, clusters)
	if err != nil {
		return nil, err
	}
	dynamicEnvs, err := loadDynamicEnvironments(yamlEnvs, sb)
	if err != nil {
		return nil, err
	}

	merged := make(map[string]terra.Environment)
	for k, e := range yamlEnvs {
		merged[k] = e
	}
	for k, e := range dynamicEnvs {
		merged[k] = e
	}

	return merged, nil
}

func loadDynamicEnvironments(yamlEnvironments map[string]terra.Environment, sb statebucket.StateBucket) (map[string]terra.Environment, error) {
	dynamicEnvironments, err := sb.Environments()
	if err != nil {
		return nil, err
	}

	result := make(map[string]terra.Environment)
	for _, dynamicEnv := range dynamicEnvironments {
		if _, exists := yamlEnvironments[dynamicEnv.Name]; exists {
			return nil, fmt.Errorf("error laoding dynamic environment %q: an environment by that name is already declared in YAML", dynamicEnv.Name)
		}
		template, exists := yamlEnvironments[dynamicEnv.Template]
		if !exists {
			return nil, fmt.Errorf("error loading dynamic environment %q: template %q is not declared in YAML", dynamicEnv.Name, dynamicEnv.Template)
		}

		var _fiab terra.Fiab
		if dynamicEnv.Hybrid {
			_fiab = terra.NewFiab(dynamicEnv.Fiab.Name, dynamicEnv.Fiab.IP)
		}

		_releases := make(map[string]terra.AppRelease)

		for _, r := range template.Releases() {
			templateRelease := r.(terra.AppRelease)
			appVersion := templateRelease.AppVersion()
			if override, exists := dynamicEnv.VersionPins[templateRelease.Name()]; exists {
				appVersion = override
			}
			_releases[templateRelease.Name()] = &appRelease{
				release: release{
					name:           templateRelease.Name(),
					releaseType:    templateRelease.Type(),
					chartVersion:   templateRelease.ChartVersion(),
					chartName:      templateRelease.ChartName(),
					repo:           templateRelease.Repo(),
					namespace:      environmentNamespace(dynamicEnv.Name), // make sure we use _this_ environment name to create the namespace, not the template name
					clusterName:    templateRelease.ClusterName(),
					clusterAddress: templateRelease.ClusterAddress(),
					destination:    nil, // replaced after env is constructed
				},
				appVersion: appVersion,
			}
		}

		env := NewEnvironment(dynamicEnv.Name, template.Base(), template.DefaultCluster(), terra.Dynamic, template.Name(), _fiab, _releases)
		result[dynamicEnv.Name] = env
		for _, r := range env.Releases() {
			r.(*appRelease).destination = env
		}
	}

	return result, nil
}

// loadYamlEnvironments scans through the environments/ subdirectory and build a slice of defined environments
func loadYamlEnvironments(configRepoPath string, versions Versions, clusters map[string]terra.Cluster) (map[string]terra.Environment, error) {
	configDir := path.Join(configRepoPath, envConfigDir)

	destConfigs, err := loadDestinationsFromDirectory(configDir, terra.EnvironmentDestination)
	if err != nil {
		return nil, err
	}

	result := make(map[string]terra.Environment)

	for _, destConfig := range destConfigs {
		if cluster, exists := clusters[destConfig.name]; exists {
			return nil, fmt.Errorf("cluster name %s conflicts with environment name %s", cluster.Name(), destConfig.name)
		}
		env, err := loadEnvironment(destConfig, versions, clusters)
		if err != nil {
			return nil, err
		}
		result[env.Name()] = env
	}

	return result, nil
}

func loadEnvironment(destConfig destinationConfig, _versions Versions, clusters map[string]terra.Cluster) (terra.Environment, error) {
	envName := destConfig.name
	envBase := destConfig.base

	log.Debug().Msgf("Loading environment %s", envName)

	var envConfig serializers.Environment
	err := yaml.Unmarshal(destConfig.mergedYaml, &envConfig)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration for environment %s: %v", envName, err)
	}

	defaultClusterName := envConfig.DefaultCluster
	if defaultClusterName == "" {
		return nil, fmt.Errorf("environment %s does not have a valid default cluster", envName)
	}
	defaultCluster, exists := clusters[defaultClusterName]
	if !exists {
		return nil, fmt.Errorf("environment %s: default cluster %s is not defined in %s directory", envName, defaultClusterName, clusterConfigDir)
	}

	lifecycle := defaultEnvironmentLifecycle
	if envConfig.Lifecycle != "" {
		if err := yaml.Unmarshal([]byte(envConfig.Lifecycle), &lifecycle); err != nil {
			return nil, fmt.Errorf("environment %s: invalid lifecycle: %v", envName, err)
		}
		if lifecycle == terra.Dynamic {
			return nil, fmt.Errorf("environment %s: environments declared in yaml files cannot have a dynamic lifecycle", envName)
		}
	}

	_releases := make(map[string]terra.AppRelease)

	for releaseName, releaseDefn := range envConfig.Releases {
		// Skip releases that aren't enabled
		if !releaseDefn.Enabled {
			continue
		}

		_cluster := defaultCluster
		if releaseDefn.Cluster != "" {
			if _, exists := clusters[releaseDefn.Cluster]; !exists {
				return nil, fmt.Errorf("environment %s: release %s: cluster %s is not defined in %s directory", envName, releaseName, releaseDefn.Cluster, clusterConfigDir)
			}
			_cluster = clusters[releaseDefn.Cluster]
		}
		clusterName := _cluster.Name()
		clusterAddress := _cluster.Address()

		// chart version
		chartVersion := releaseDefn.ChartVersion
		if chartVersion == "" {
			chartVersion = _versions.GetSnapshot(terra.AppReleaseType, versionSetFor(envName)).ChartVersion(releaseName)
		}
		if chartVersion == "" {
			chartVersion = _versions.GetSnapshot(terra.AppReleaseType, Dev).ChartVersion(releaseName)
		}
		if chartVersion == "" {
			return nil, fmt.Errorf("environment %s: could not identify chart version for release %s", envName, releaseName)
		}

		// app version
		appVersion := releaseDefn.AppVersion
		if appVersion == "" {
			appVersion = _versions.GetSnapshot(terra.AppReleaseType, versionSetFor(envName)).AppVersion(releaseName)
		}
		if appVersion == "" {
			appVersion = _versions.GetSnapshot(terra.AppReleaseType, Dev).AppVersion(releaseName)
		}
		if appVersion == "" {
			return nil, fmt.Errorf("environment %s: could not identify app version for release %s", envName, releaseName)
		}

		// namespace
		// eg. terra-dev
		namespace := environmentNamespace(envName)

		// chartName
		chartName := releaseDefn.ChartName
		if chartName == "" {
			// chart name defaults to release name if it is not set
			chartName = releaseName
		}

		// repo
		repo := releaseDefn.Repo
		if repo == "" {
			repo = defaultChartRepo
		}

		_release := &appRelease{
			appVersion: appVersion,
			release: release{
				name:           releaseName,
				releaseType:    terra.AppReleaseType,
				chartVersion:   chartVersion,
				chartName:      chartName,
				repo:           repo,
				namespace:      namespace,
				clusterName:    clusterName,
				clusterAddress: clusterAddress,
			},
		}

		_releases[releaseName] = _release
	}

	log.Debug().Msgf("Found %d releases for environment %s", len(_releases), envName)

	env := NewEnvironment(
		envName,
		envBase,
		defaultClusterName,
		lifecycle,
		"",
		nil,
		_releases,
	)

	for _, _release := range _releases {
		_release.(*appRelease).destination = env
	}

	return env, nil
}

// loadClusters scans through the cluster/ subdirectory and build a slice of defined clusters
func loadClusters(configRepoPath string, versions Versions) (map[string]terra.Cluster, error) {
	configDir := path.Join(configRepoPath, clusterConfigDir)

	destConfigs, err := loadDestinationsFromDirectory(configDir, terra.ClusterDestination)
	if err != nil {
		return nil, err
	}

	result := make(map[string]terra.Cluster)

	for _, destConfig := range destConfigs {
		_cluster, err := loadCluster(destConfig, versions)
		if err != nil {
			return nil, err
		}
		result[_cluster.Name()] = _cluster
	}

	return result, nil
}

func loadCluster(destConfig destinationConfig, _versions Versions) (terra.Cluster, error) {
	clusterName := destConfig.name
	clusterBase := destConfig.base

	var clusterDefn serializers.Cluster
	err := yaml.Unmarshal(destConfig.mergedYaml, &clusterDefn)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration for cluster %s: %v", clusterName, err)
	}

	clusterAddress := clusterDefn.Address
	if clusterAddress == "" {
		return nil, fmt.Errorf("cluster %s does not have a valid API address, please set `address` key in config file", clusterName)
	}

	releases := make(map[string]terra.ClusterRelease)

	for releaseName, releaseDefn := range clusterDefn.Releases {
		// Skip releaes that aren't enabled
		if !releaseDefn.Enabled {
			continue
		}

		// Release is enabled, so configure with proper settings
		// chart version
		chartVersion := releaseDefn.ChartVersion
		if chartVersion == "" {
			chartVersion = _versions.GetSnapshot(terra.ClusterReleaseType, versionSetFor(clusterName)).ChartVersion(releaseName)
		}
		if chartVersion == "" {
			chartVersion = _versions.GetSnapshot(terra.ClusterReleaseType, Dev).ChartVersion(releaseName)
		}
		if chartVersion == "" {
			return nil, fmt.Errorf("cluster %s: could not identify chart version for release %s", clusterName, releaseName)
		}

		// namespace
		namespace := releaseDefn.Namespace
		if namespace == "" {
			return nil, fmt.Errorf("cluster %s: release %s does not have a valid namespace", clusterName, releaseName)
		}

		// chartName
		chartName := releaseDefn.ChartName
		if chartName == "" {
			// chart name defaults to release name if it is not set
			chartName = releaseName
		}

		// repo
		repo := releaseDefn.Repo
		if repo == "" {
			repo = defaultChartRepo
		}

		releases[releaseName] = &clusterRelease{
			release: release{
				name:           releaseName,
				releaseType:    terra.ClusterReleaseType,
				chartVersion:   chartVersion,
				chartName:      chartName,
				repo:           repo,
				namespace:      namespace,
				clusterName:    clusterName,
				clusterAddress: clusterAddress,
			},
		}
	}

	_cluster := NewCluster(
		clusterName,
		clusterBase,
		clusterDefn.Address,
		releases,
	)

	for _, _release := range releases {
		_release.(*clusterRelease).destination = _cluster
	}

	return _cluster, nil
}

// Silly heuristic... If the destination name ends with "alpha", use the alpha snapshot, etc.
// Defaults to dev snapshot.
func versionSetFor(destinationName string) VersionSet {
	for _, versionSet := range VersionSets() {
		if strings.HasSuffix(destinationName, versionSet.String()) {
			return versionSet
		}
	}

	return Dev
}

type destinationConfig struct {
	name       string
	base       string
	mergedYaml []byte
}

// loadDestinationsFromDirectory loads the set of configured clusters or environments from a config directory
func loadDestinationsFromDirectory(configDir string, destType terra.DestinationType) (map[string]destinationConfig, error) {
	destConfigs := make(map[string]destinationConfig)

	if _, err := os.Stat(configDir); err != nil {
		return nil, fmt.Errorf("%s config directory does not exist: %s", destType, configDir)
	}

	matches, err := filepath.Glob(path.Join(configDir, "*", withYamlSuffix("*")))
	if err != nil {
		return nil, fmt.Errorf("error loading %s configs from %s: %v", destType, configDir, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no %s configs found in %s", destType, configDir)
	}

	for _, filename := range matches {
		base := path.Base(path.Dir(filename))
		name := strings.TrimSuffix(path.Base(filename), yamlSuffix)

		if conflict, ok := destConfigs[name]; ok {
			return nil, fmt.Errorf("%s name conflict %s (%s) and %s (%s)", destType, name, base, conflict.name, conflict.base)
		}

		mergedYaml, err := mergeDestinationYaml(configDir, base, name)
		if err != nil {
			return nil, fmt.Errorf("error loading YAML config for %s %s: %v", destType, name, err)
		}

		destConfig := destinationConfig{
			name:       name,
			base:       base,
			mergedYaml: mergedYaml,
		}

		destConfigs[name] = destConfig
	}

	return destConfigs, nil
}

func mergeDestinationYaml(configDir string, base string, name string) ([]byte, error) {
	defaultsFile := path.Join(configDir, defaultsFileName)              // eg. environments/defaults.yaml
	baseFile := path.Join(configDir, withYamlSuffix(base))              // eg. environments/live.yaml
	destinationFile := path.Join(configDir, base, withYamlSuffix(name)) // eg. environments/live/dev.yaml

	return deepmerge.Merge(defaultsFile, baseFile, destinationFile)
}

func withYamlSuffix(baseName string) string {
	return fmt.Sprintf("%s%s", baseName, yamlSuffix)
}
