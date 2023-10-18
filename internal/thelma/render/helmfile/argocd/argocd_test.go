package argocd

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteDestinationValuesFile(t *testing.T) {
	t.Skip("TODO")
}

func Test_argoDestinationsForEnvironment(t *testing.T) {
	tests := []struct {
		name                      string
		releaseMockConfigurations []func(m *mocks.AppRelease)
		want                      []ArgoDestination
	}{
		{
			name: "simple case",
			releaseMockConfigurations: []func(m *mocks.AppRelease){
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.4").Once()
					m.EXPECT().Namespace().Return("namespace").Once()
				},
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.4").Once()
					m.EXPECT().Namespace().Return("namespace").Once()
				},
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.4").Once()
					m.EXPECT().Namespace().Return("namespace").Once()
				},
			},
			want: []ArgoDestination{
				{
					Server:    "192.168.3.4",
					Namespace: "namespace",
				},
			},
		},
		{
			name: "multi cluster",
			releaseMockConfigurations: []func(m *mocks.AppRelease){
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.4").Once()
					m.EXPECT().Namespace().Return("namespace").Once()
				},
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.5").Once()
					m.EXPECT().Namespace().Return("namespace").Once()
				},
			},
			want: []ArgoDestination{
				{
					Server:    "192.168.3.4",
					Namespace: "namespace",
				},
				{
					Server:    "192.168.3.5",
					Namespace: "namespace",
				},
			},
		},
		{
			name: "multi namespace",
			releaseMockConfigurations: []func(m *mocks.AppRelease){
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.4").Once()
					m.EXPECT().Namespace().Return("namespace-1").Once()
				},
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.4").Once()
					m.EXPECT().Namespace().Return("namespace-2").Once()
				},
			},
			want: []ArgoDestination{
				{
					Server:    "192.168.3.4",
					Namespace: "namespace-1",
				},
				{
					Server:    "192.168.3.4",
					Namespace: "namespace-2",
				},
			},
		},
		{
			name: "multi namespace and multi cluster",
			releaseMockConfigurations: []func(m *mocks.AppRelease){
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.4").Once()
					m.EXPECT().Namespace().Return("namespace-1").Once()
				},
				func(m *mocks.AppRelease) {
					m.EXPECT().ClusterAddress().Return("192.168.3.5").Once()
					m.EXPECT().Namespace().Return("namespace-2").Once()
				},
			},
			want: []ArgoDestination{
				{
					Server:    "192.168.3.4",
					Namespace: "namespace-1",
				},
				{
					Server:    "192.168.3.5",
					Namespace: "namespace-2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockReleases []terra.Release
			for _, configFn := range tt.releaseMockConfigurations {
				mockRelease := mocks.NewAppRelease(t)
				configFn(mockRelease)
				mockReleases = append(mockReleases, mockRelease)
			}
			mockEnvironment := mocks.NewEnvironment(t)
			mockEnvironment.EXPECT().Releases().Return(mockReleases).Once()
			if got := argoDestinationsForEnvironment(mockEnvironment); !assert.ElementsMatch(t, got, tt.want) {
				t.Errorf("argoDestinationsForEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}
