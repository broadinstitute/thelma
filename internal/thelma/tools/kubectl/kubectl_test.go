package kubectl

import (
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	statemocks "github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CreateAndDeleteNamespace(t *testing.T) {
	mockKubeCfg := mocks.NewKubeconfig(t)
	mockRunner := shell.DefaultMockRunner()
	_kubectl := New(mockRunner, mockKubeCfg)

	env := statemocks.NewEnvironment(t)
	env.EXPECT().Namespace().Return("terra-fake-bee")
	env.EXPECT().Lifecycle().Return(terra.Dynamic)

	ktx := mocks.NewKubectx(t)
	ktx.EXPECT().Namespace().Return("terra-fake-bee")
	ktx.EXPECT().ContextName().Return("my-ctx")

	mockKubeCfg.EXPECT().ConfigFile().Return("fake-kubecfg")
	mockKubeCfg.EXPECT().ForEnvironment(env).Return([]kubecfg.Kubectx{ktx}, nil)

	mockRunner.ExpectCmd(shell.Command{
		Prog: "kubectl",
		Args: []string{"--context", "my-ctx", "--namespace", "terra-fake-bee", "create", "namespace", "terra-fake-bee"},
		Env:  []string{"KUBECONFIG=fake-kubecfg"},
	})
	require.NoError(t, _kubectl.CreateNamespace(env))

	mockRunner.ExpectCmd(shell.Command{
		Prog: "kubectl",
		Args: []string{"--context", "my-ctx", "--namespace", "terra-fake-bee", "delete", "namespace", "--ignore-not-found", "terra-fake-bee"},
		Env:  []string{"KUBECONFIG=fake-kubecfg"},
	})
	require.NoError(t, _kubectl.DeleteNamespace(env))
}

func Test_Shutdown(t *testing.T) {
	mockKubeCfg := mocks.NewKubeconfig(t)
	mockRunner := shell.DefaultMockRunner()
	_kubectl := New(mockRunner, mockKubeCfg)

	env := statemocks.NewEnvironment(t)
	env.EXPECT().Name().Return("staging")

	ktx1 := mocks.NewKubectx(t)
	ktx1.EXPECT().Namespace().Return("terra-staging")
	ktx1.EXPECT().ContextName().Return("my-ctx-1")

	ktx2 := mocks.NewKubectx(t)
	ktx2.EXPECT().Namespace().Return("terra-staging")
	ktx2.EXPECT().ContextName().Return("my-ctx-2")

	mockKubeCfg.EXPECT().ConfigFile().Return("fake-kubecfg")
	mockKubeCfg.EXPECT().ForEnvironment(env).Return([]kubecfg.Kubectx{ktx1, ktx2}, nil)

	mockRunner.ExpectCmd(shell.Command{
		Prog: "kubectl",
		Args: []string{"--context", "my-ctx-1", "--namespace", "terra-staging", "scale", "--replicas=0", "deployment", "--all"},
		Env:  []string{"KUBECONFIG=fake-kubecfg"},
	})
	mockRunner.ExpectCmd(shell.Command{
		Prog: "kubectl",
		Args: []string{"--context", "my-ctx-2", "--namespace", "terra-staging", "scale", "--replicas=0", "deployment", "--all"},
		Env:  []string{"KUBECONFIG=fake-kubecfg"},
	})
	mockRunner.ExpectCmd(shell.Command{
		Prog: "kubectl",
		Args: []string{"--context", "my-ctx-1", "--namespace", "terra-staging", "scale", "--replicas=0", "statefulset", "--all"},
		Env:  []string{"KUBECONFIG=fake-kubecfg"},
	})
	mockRunner.ExpectCmd(shell.Command{
		Prog: "kubectl",
		Args: []string{"--context", "my-ctx-2", "--namespace", "terra-staging", "scale", "--replicas=0", "statefulset", "--all"},
		Env:  []string{"KUBECONFIG=fake-kubecfg"},
	})

	require.NoError(t, _kubectl.ShutDown(env))
}

func Test_DeletePVCs(t *testing.T) {
	mockKubeCfg := mocks.NewKubeconfig(t)
	mockRunner := shell.DefaultMockRunner()
	_kubectl := New(mockRunner, mockKubeCfg)

	liveEnv := statemocks.NewEnvironment(t)
	liveEnv.EXPECT().Lifecycle().Return(terra.Static)

	assert.ErrorContains(t, _kubectl.DeletePVCs(liveEnv), "can only be called for dynamic environments")

	env := statemocks.NewEnvironment(t)
	env.EXPECT().Name().Return("fake-bee")
	env.EXPECT().Lifecycle().Return(terra.Dynamic)

	ktx := mocks.NewKubectx(t)
	ktx.EXPECT().Namespace().Return("terra-fake-bee")
	ktx.EXPECT().ContextName().Return("my-ctx")

	mockKubeCfg.EXPECT().ConfigFile().Return("fake-kubecfg")
	mockKubeCfg.EXPECT().ForEnvironment(env).Return([]kubecfg.Kubectx{ktx}, nil)

	mockRunner.ExpectCmd(shell.Command{
		Prog: "kubectl",
		Args: []string{"--context", "my-ctx", "--namespace", "terra-fake-bee", "delete", "persistentvolumeclaims", "--all", "--wait=true"},
		Env:  []string{"KUBECONFIG=fake-kubecfg"},
	})

	require.NoError(t, _kubectl.DeletePVCs(env))
}

func Test_parsePortFromPortForwardOutput(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want int
	}{
		{
			name: "port when normal",
			arg: `
Forwarding from 127.0.0.1:58795 -> 5432
Forwarding from [::1]:58795 -> 5432

`,
			want: 58795,
		},
		{
			name: "zero when nothing",
			arg:  "",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePortFromPortForwardOutput(tt.arg); got != tt.want {
				t.Errorf("parsePortFromPortForwardOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
