package logs

import (
	"context"
	k8sclients "github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes"
	"github.com/broadinstitute/thelma/internal/thelma/clients/kubernetes/kubecfg"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
)

func newResourceCache(k8sclients k8sclients.Clients) *resourceCache {
	return &resourceCache{
		cache:      make(map[string]cacheEntry),
		mutex:      sync.Mutex{},
		k8sclients: k8sclients,
	}
}

type resourceCache struct {
	cache      map[string]cacheEntry
	mutex      sync.Mutex
	k8sclients k8sclients.Clients
}

// cacheEntry contains all the deployments and statefulsets in a given Kubectx (cluster + namespace)
type cacheEntry struct {
	deployments  []appsv1.Deployment
	statefulsets []appsv1.StatefulSet
}

func (r *resourceCache) get(kubectx kubecfg.Kubectx) (cacheEntry, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	entry, exists := r.cache[kubectx.ContextName()]
	if exists {
		return entry, nil
	}

	apiClient, err := r.k8sclients.ForKubectx(kubectx)
	if err != nil {
		return entry, err
	}

	deployments, err := apiClient.AppsV1().Deployments(kubectx.Namespace()).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return entry, err
	}
	entry.deployments = deployments.Items

	statefulsets, err := apiClient.AppsV1().StatefulSets(kubectx.Namespace()).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return entry, err
	}
	entry.statefulsets = statefulsets.Items

	r.cache[kubectx.ContextName()] = entry
	return entry, nil
}
