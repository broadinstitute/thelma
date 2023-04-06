package testing

import "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/testing/mocks"

type KubeMocks struct {
	KubeClient   *mocks.KubeClient
	CoreV1       *mocks.CoreV1
	ConfigMaps   *mocks.ConfigMaps
	Pods         *mocks.Pods
	Secrets      *mocks.Secrets
	Services     *mocks.Services
	AppsV1       *mocks.AppsV1
	Deployments  *mocks.Deployments
	StatefulSets *mocks.StatefulSets
}

func NewKubeMocks(namespace string) *KubeMocks {
	client := &mocks.KubeClient{}

	corev1 := &mocks.CoreV1{}
	client.EXPECT().CoreV1().Return(corev1)
	cms := &mocks.ConfigMaps{}
	corev1.EXPECT().ConfigMaps(namespace).Return(cms)
	pods := &mocks.Pods{}
	corev1.EXPECT().Pods(namespace).Return(pods)
	secrets := &mocks.Secrets{}
	corev1.EXPECT().Secrets(namespace).Return(secrets)
	services := &mocks.Services{}
	corev1.EXPECT().Services(namespace).Return(services)

	appsv1 := &mocks.AppsV1{}
	client.EXPECT().AppsV1().Return(appsv1)
	deployments := &mocks.Deployments{}
	appsv1.EXPECT().Deployments(namespace).Return(deployments)
	statefulsets := &mocks.StatefulSets{}
	appsv1.EXPECT().StatefulSets(namespace).Return(statefulsets)

	return &KubeMocks{
		KubeClient:   client,
		CoreV1:       corev1,
		ConfigMaps:   cms,
		Pods:         pods,
		Secrets:      secrets,
		Services:     services,
		AppsV1:       appsv1,
		Deployments:  deployments,
		StatefulSets: statefulsets,
	}
}
