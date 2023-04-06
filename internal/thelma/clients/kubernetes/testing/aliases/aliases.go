package aliases

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type KubeClient interface {
	kubernetes.Interface
}

type CoreV1 interface {
	corev1.CoreV1Interface
}

type Secrets interface {
	corev1.SecretInterface
}

type ConfigMaps interface {
	corev1.ConfigMapInterface
}

type Pods interface {
	corev1.PodInterface
}

type Services interface {
	corev1.ServiceInterface
}

type AppsV1 interface {
	appsv1.AppsV1Interface
}

type StatefulSets interface {
	appsv1.StatefulSetInterface
}

type Deployments interface {
	appsv1.DeploymentInterface
}

type Watch interface {
	watch.Interface
}
