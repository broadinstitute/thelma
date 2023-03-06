package render

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"regexp"
	"sort"
	"testing"

	"github.com/broadinstitute/thelma/internal/thelma/app/builder"
	"github.com/broadinstitute/thelma/internal/thelma/cli"
	"github.com/broadinstitute/thelma/internal/thelma/render"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
	"github.com/broadinstitute/thelma/internal/thelma/render/resolver"
	"github.com/broadinstitute/thelma/internal/thelma/render/scope"
	"github.com/broadinstitute/thelma/internal/thelma/render/validator"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/filter"
	"github.com/broadinstitute/thelma/internal/thelma/state/testing/statefixtures"
	. "github.com/broadinstitute/thelma/internal/thelma/utils/testutils"
	"github.com/stretchr/testify/assert"
)

const stateFixture = statefixtures.Default

// TestRenderArgParsing Given given a set of CLI args, verify that options structures are populated correctly
func TestRenderArgParsing(t *testing.T) {
	type expectedAttrs struct {
		renderOptions *render.Options
		helmfileArgs  *helmfile.Args
	}
	type testConfig struct {
		t        *testing.T
		options  *cli.Options
		expected *expectedAttrs
	}

	//nolint:staticcheck // SA1019
	fixture, err := statefixtures.LoadFixture(stateFixture)
	require.NoError(t, err)

	testCases := []struct {
		description   string                     // testcase description
		arguments     []string                   // renderCLI args to pass in
		setupFn       func(tc *testConfig) error // optional hook for extra setup
		expectedError *regexp.Regexp             // expected error
	}{
		{
			description:   "invalid argument",
			arguments:     Args("render --foo"),
			expectedError: regexp.MustCompile("unknown flag"),
		},
		{
			description:   "no arguments",
			arguments:     Args("render"),
			expectedError: regexp.MustCompile("please specify at least one release"),
		},
		{
			description:   "positional and -r cannot be combined",
			arguments:     Args("render -r foo foo"),
			expectedError: regexp.MustCompile(`releases can either be specified with the --release flag or via positional argument, not both`),
		},
		{
			description:   "unknown release",
			arguments:     Args("render -r foo"),
			expectedError: regexp.MustCompile(`--release: unknown release\(s\) foo`),
		},
		{
			description:   "unknown multiple releases",
			arguments:     Args("render -r foo,bar,leonardo -r sam,baz"),
			expectedError: regexp.MustCompile(`--release: unknown release\(s\) bar, baz, foo`),
		},
		{
			description:   "unknown environment",
			arguments:     Args("render -e foo ALL"),
			expectedError: regexp.MustCompile(`--environment: unknown environment\(s\) foo`),
		},
		{
			description:   "unknown cluster",
			arguments:     Args("render -c foo ALL"),
			expectedError: regexp.MustCompile(`--cluster: unknown cluster\(s\) foo`),
		},
		{
			description:   "unknown destination type",
			arguments:     Args("render --destination-type foo ALL"),
			expectedError: regexp.MustCompile(`--destination-type: unknown destination-type\(s\) foo`),
		},
		{
			description:   "unknown destination base",
			arguments:     Args("render --destination-base foo ALL"),
			expectedError: regexp.MustCompile(`--destination-base: unknown destination-base\(s\) foo`),
		},
		{
			description:   "unknown environment lifecycle",
			arguments:     Args("render --environment-lifecycle foo ALL"),
			expectedError: regexp.MustCompile(`--environment-lifecycle: unknown environment-lifecycle\(s\) foo`),
		},
		{
			description:   "unknown environment template",
			arguments:     Args("render --environment-template foo ALL"),
			expectedError: regexp.MustCompile(`--environment-template: unknown environment-template\(s\) foo`),
		},
		{
			description:   "-e/-c can't be combined with destination filters",
			arguments:     Args("render -e dev --environment-template swatomation ALL"),
			expectedError: regexp.MustCompile(`--environment cannot be combined with --environment-template`),
		},
		{
			description:   "no releases match arguments",
			arguments:     Args("render -c terra-dev sam"),
			expectedError: regexp.MustCompile(`0 releases matched command-line arguments`),
		},
		{
			description:   "-a and -r cannot be combined",
			arguments:     Args("render -r leonardo -a cromwell"),
			expectedError: regexp.MustCompile(`one or the other but not both`),
		},
		{
			description:   "--app-version should require single chart",
			arguments:     Args("render --app-version 1.0.0 -r leonardo,cromwell"),
			expectedError: regexp.MustCompile("cannot be used with selectors that match multiple charts"),
		},
		{
			description:   "--chart-version should require single chart",
			arguments:     Args("render --chart-version 1.0.0 ALL"),
			expectedError: regexp.MustCompile("cannot be used with selectors that match multiple charts"),
		},
		{
			description:   "--values-file should require -r",
			arguments:     Args("render --values-file %s -r workspacemanager -r agora", path.Join(t.TempDir(), "does-not-exist.yaml")),
			expectedError: regexp.MustCompile("cannot be used with selectors that match multiple charts"),
		},
		{
			description:   "--values-file must exist",
			arguments:     Args("render -e dev -r leonardo --values-file %s", path.Join(t.TempDir(), "does-not-exist.yaml")),
			expectedError: regexp.MustCompile("values file does not exist: .*/does-not-exist.yaml"),
		},
		{
			description:   "--chart-dir and --chart-version incompatible",
			arguments:     Args("render -e dev -r leonardo --chart-dir %s --chart-version 1.0.0", t.TempDir()),
			expectedError: regexp.MustCompile("only one of --chart-dir or --chart-version may be specified"),
		},
		{
			description:   "--chart-dir must exist",
			arguments:     Args("render -e dev -r leonardo --chart-dir chart/dir/does/not/exist"),
			expectedError: regexp.MustCompile("chart dir does not exist: .*chart/dir/does/not/exist"),
		},
		{
			description:   "--argocd and --app-version incompatible",
			arguments:     Args("render -e dev -r leonardo --app-version 1.0.0 --argocd"),
			expectedError: regexp.MustCompile("--argocd cannot be used with.*--app-version"),
		},
		{
			description:   "--argocd and --chart-version incompatible",
			arguments:     Args("render -e dev -r leonardo --chart-version 1.0.0 --argocd"),
			expectedError: regexp.MustCompile("--argocd cannot be used with.*--chart-version"),
		},
		{
			description:   "--argocd and --chart-dir incompatible",
			arguments:     Args("render -e dev -r leonardo --chart-dir=%s --argocd", t.TempDir()),
			expectedError: regexp.MustCompile("--argocd cannot be used with.*--chart-dir"),
		},
		{
			description:   "--argocd and --values-file incompatible",
			arguments:     Args("render -e dev -r leonardo --values-file=%s --argocd", "does-not-exist.yaml"),
			expectedError: regexp.MustCompile("--argocd cannot be used with.*--values-file"),
		},
		{
			description:   "--stdout and --output-dir incompatible",
			arguments:     Args("render -e dev -r leonardo -d /tmp/output --stdout"),
			expectedError: regexp.MustCompile("--stdout cannot be used with --output-dir"),
		},
		{
			description:   "--parallel-workers and --stdout incompatible",
			arguments:     Args("render --parallel-workers 10 --stdout ALL"),
			expectedError: regexp.MustCompile("--parallel-workers cannot be used with --stdout"),
		},
		{
			description:   "--cluster and --app-version incompatible",
			arguments:     Args("render --cluster terra-perf -r yale --app-version=0.0.1"),
			expectedError: regexp.MustCompile("--app-version cannot be used for cluster releases"),
		},
		{
			description:   "--exact-release ALL invalid",
			arguments:     Args("render --exact-release ALL"),
			expectedError: regexp.MustCompile("--exact-release cannot be used with ALL"),
		},
		{
			description:   "--validate invalid arg",
			arguments:     Args("render --validate foo ALL"),
			expectedError: regexp.MustCompile("--validate: invalid validate mode.*"),
		},
		{
			description: "config repo path must be set",
			arguments:   []string{"render"},
			setupFn: func(tc *testConfig) error {
				tc.options.ConfigureThelma(func(b builder.ThelmaBuilder) {
					b.SetHome("")
				})
				return nil
			},
			expectedError: regexp.MustCompile("please specify path to terra-helmfile clone"),
		},
		{
			description: "-r should set release name",
			setupFn: func(tc *testConfig) error {
				tc.options.SetArgs(Args("render -r datarepo"))
				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("datarepo", "alpha"),
					fixture.Release("datarepo", "staging"),
					fixture.Release("datarepo", "prod"),
				}
				return nil
			},
		},
		{
			description: "first positional should set release name",
			setupFn: func(tc *testConfig) error {
				tc.options.SetArgs(Args("render -r datarepo"))
				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("datarepo", "alpha"),
					fixture.Release("datarepo", "staging"),
					fixture.Release("datarepo", "prod"),
				}
				return nil
			},
		},
		{
			description: "-e should set environment",
			setupFn: func(tc *testConfig) error {
				tc.options.SetArgs(Args("render -e dev ALL"))
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("agora", "dev"),
					fixture.Release("buffer", "dev"),
					fixture.Release("cromwell", "dev"),
					fixture.Release("externalcreds", "dev"),
					fixture.Release("leonardo", "dev"),
					fixture.Release("rawls", "dev"),
					fixture.Release("sam", "dev"),
					fixture.Release("workspacemanager", "dev"),
				}
				return nil
			},
		},
		{
			description: "-e should work for dynamic environments",
			setupFn: func(tc *testConfig) error {
				tc.options.SetArgs(Args("render -e fiab-nerdy-walrus ALL"))
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("agora", "fiab-nerdy-walrus"),
					fixture.Release("leonardo", "fiab-nerdy-walrus"),
					fixture.Release("sam", "fiab-nerdy-walrus"),
					fixture.Release("rawls", "fiab-nerdy-walrus"),
					fixture.Release("opendj", "fiab-nerdy-walrus"),
					fixture.Release("workspacemanager", "fiab-nerdy-walrus"),
				}
				return nil
			},
		},
		{
			description: "-e and -r should intersect",
			setupFn: func(tc *testConfig) error {
				tc.options.SetArgs(Args("render -e dev -r externalcreds"))
				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("externalcreds", "dev"),
				}
				return nil
			},
		},
		{
			description: "multiple -e and -r flags should be additive",
			setupFn: func(tc *testConfig) error {
				tc.options.SetArgs(Args("render -e dev -e perf -e alpha -r sam -r leonardo"))
				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("sam", "dev"),
					fixture.Release("sam", "perf"),
					fixture.Release("sam", "alpha"),
					fixture.Release("leonardo", "dev"),
					fixture.Release("leonardo", "perf"),
					fixture.Release("leonardo", "alpha"),
				}
				return nil
			},
		},
		{
			description: "-c should set cluster",
			setupFn: func(tc *testConfig) error {
				tc.options.SetArgs(Args("render -c terra-alpha ALL"))
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("diskmanager", "terra-alpha"),
					fixture.Release("install-secrets-manager", "terra-alpha"),
					fixture.Release("terra-prometheus", "terra-alpha"),
					fixture.Release("yale", "terra-alpha"),
				}
				return nil
			},
		},
		{
			description: "-d should set output directory",
			setupFn: func(tc *testConfig) error {
				dir := tc.t.TempDir()
				tc.options.SetArgs(Args("render -d %s ALL", dir))
				tc.expected.renderOptions.OutputDir = dir
				return nil
			},
		},
		{
			description: "--stdout should set stdout",
			arguments:   Args("render --stdout ALL"),
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.Stdout = true
				return nil
			},
		},
		{
			description: "--parallel-workers should set workers",
			arguments:   Args("render --parallel-workers 32 ALL"),
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.ParallelWorkers = 32
				return nil
			},
		},
		{
			description: "--app-version should set app version",
			arguments:   Args("render -e dev -r leonardo --app-version 1.2.3"),
			setupFn: func(tc *testConfig) error {
				version := "1.2.3"
				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("leonardo", "dev"),
				}
				tc.expected.helmfileArgs.AppVersion = &version
				return nil
			},
		},
		{
			description: "--chart-version should set chart version",
			arguments:   Args("render -e dev leonardo --chart-version 4.5.6"),
			setupFn: func(tc *testConfig) error {
				version := "4.5.6"
				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("leonardo", "dev"),
				}
				tc.expected.helmfileArgs.ChartVersion = &version
				return nil
			},
		},
		{
			description: "--chart-dir should set chart source dir",
			setupFn: func(tc *testConfig) error {
				chartDir := tc.t.TempDir()
				tc.expected.renderOptions.ChartSourceDir = chartDir
				tc.options.SetArgs(Args("render --chart-dir %s ALL", chartDir))
				return nil
			},
		},
		{
			description: "--mode=development should set mode to development",
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.ResolverMode = resolver.Development
				tc.options.SetArgs(Args("render --mode development ALL"))
				return nil
			},
		},
		{
			description: "--mode=deploy should set mode to deploy",
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.ResolverMode = resolver.Deploy
				tc.options.SetArgs(Args("render --mode deploy ALL"))
				return nil
			},
		},
		{
			description: "--values-file once should set single values file",
			setupFn: func(tc *testConfig) error {
				valuesDir := tc.t.TempDir()
				valuesFile := path.Join(valuesDir, "v1.yaml")
				if err := os.WriteFile(valuesFile, []byte("# fake values file"), 0644); err != nil {
					return err
				}

				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("leonardo", "dev"),
				}

				tc.expected.helmfileArgs.ValuesFiles = []string{valuesFile}

				tc.options.SetArgs(Args("render -e dev -r leonardo --values-file %s", valuesFile))

				return nil
			},
		},
		{
			description: "--values-file multiple times should set multiple values files",
			setupFn: func(tc *testConfig) error {
				valuesDir := tc.t.TempDir()
				valuesFiles := []string{
					path.Join(valuesDir, "v1.yaml"),
					path.Join(valuesDir, "v2.yaml"),
					path.Join(valuesDir, "v3.yaml"),
				}
				for _, f := range valuesFiles {
					if err := os.WriteFile(f, []byte("# fake values file"), 0644); err != nil {
						return err
					}
				}

				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("leonardo", "dev"),
				}
				tc.expected.helmfileArgs.ValuesFiles = valuesFiles

				tc.options.SetArgs(Args("render -e dev -r leonardo --values-file %s --values-file %s --values-file %s", valuesFiles[0], valuesFiles[1], valuesFiles[2]))

				return nil
			},
		},
		{
			description: "--argocd should enable argocd mode",
			arguments:   Args("render --argocd ALL"),
			setupFn: func(tc *testConfig) error {
				tc.expected.helmfileArgs.ArgocdMode = true
				return nil
			},
		},
		{
			description: "--exact-release should match full name",
			arguments:   Args("render --exact-release leonardo-dev"),
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.Scope = scope.Release
				tc.expected.renderOptions.Releases = []terra.Release{
					fixture.Release("leonardo", "dev"),
				}
				return nil
			},
		},
		{
			description: "--validate should default to skip",
			arguments:   Args("render ALL"),
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.Validate = validator.Skip
				return nil
			},
		},
		{
			description: "--validate with explicit skip",
			arguments:   Args("render --validate skip ALL"),
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.Validate = validator.Skip
				return nil
			},
		},
		{
			description: "--validate with warn mode",
			arguments:   Args("render --validate warn ALL"),
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.Validate = validator.Warn
				return nil
			},
		},
		{
			description: "--validate with fail mode",
			arguments:   Args("render --validate fail ALL"),
			setupFn: func(tc *testConfig) error {
				tc.expected.renderOptions.Validate = validator.Fail
				return nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			options := cli.DefaultOptions()

			// Replace render's RunE with a noop function,
			// We're just testing argument parsing, so only test pre-/post- run hooks here
			options.SkipRun(true)

			expected := &expectedAttrs{
				renderOptions: &render.Options{},
				helmfileArgs:  &helmfile.Args{},
			}

			thelmaHome := t.TempDir()
			// set config repo path to a tmp dir we control
			options.ConfigureThelma(func(b builder.ThelmaBuilder) {
				b.WithTestDefaults(t)
				b.UseCustomStateLoader(fixture.Mocks().StateLoader)
				b.SetHome(thelmaHome)
			})

			// we expect our CLI code to populate these defaults, but users can override them in setupFn
			expected.renderOptions.OutputDir = path.Join(thelmaHome, "output")
			expected.renderOptions.ChartSourceDir = path.Join(thelmaHome, "charts")
			expected.renderOptions.ParallelWorkers = 1
			expected.renderOptions.Scope = scope.All
			expected.renderOptions.Releases = defaultReleases(fixture)

			// set command-line args
			options.SetArgs(testCase.arguments)

			// load fixture
			tc := &testConfig{
				t:        t,
				options:  options,
				expected: expected,
			}

			// call test case setupFn if defined
			if testCase.setupFn != nil {
				if err := testCase.setupFn(tc); err != nil {
					t.Errorf("setup function returned an error: %v", err)
					return
				}
			}

			// execute the test parsing code
			cmd := NewRenderCommand().(*renderCommand)
			options.AddCommand("render", cmd)
			err := cli.NewWithOptions(options).Execute()

			// if error was expected, check it
			if testCase.expectedError != nil {
				if !assert.Error(t, err, "Expected error matching %v", testCase.expectedError) {
					return
				}
				assert.Regexp(t, testCase.expectedError, err.Error())
				return
			}

			// make sure no error was returned
			if !assert.NoError(t, err, fmt.Errorf("renderCLI.execute() returned unexpected error: %v", err)) {
				return
			}

			// sort expected releases before comparison
			expectedReleases := releasesToFullNames(expected.renderOptions.Releases)
			actualReleases := releasesToFullNames(cmd.renderOptions.Releases)
			sort.Strings(expectedReleases)
			sort.Strings(actualReleases)

			assert.Equal(t, expectedReleases, actualReleases)

			expected.renderOptions.Releases = nil
			cmd.renderOptions.Releases = nil

			assert.Equal(t, expected.renderOptions, cmd.renderOptions)
			assert.Equal(t, expected.helmfileArgs, cmd.helmfileArgs)
		})
	}
}

func releasesToFullNames(releases []terra.Release) []string {
	var fullNames []string
	for _, r := range releases {
		fullNames = append(fullNames, r.FullName())
	}
	return fullNames
}

// return the default set of releases that should be matched when no filter flags are applied
func defaultReleases(fixture statefixtures.Fixture) []terra.Release {
	f := filter.Releases().DestinationMatches(
		filter.Destinations().IsCluster().Or(
			filter.Destinations().IsEnvironmentMatching(filter.Environments().HasLifecycleName("static", "template"))),
	)
	return f.Filter(fixture.AllReleases())
}
