package kubernetes

import (
	"context"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/argocd"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox/kubectl"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
	"strings"
)

type features struct {
	dbms            api.DBMS
	statefulsetName string
	container       kubectl.Container
	serviceName     string
	serviceHostName string
}

func detectFeatures(r terra.Release, k8sclient k8sclient.Interface) (*features, error) {
	var f features

	// identify statefulset
	appSelector := argocd.ApplicationSelector(argocd.ApplicationName(r))
	statefulsets, err := k8sclient.AppsV1().StatefulSets(r.Namespace()).List(
		context.Background(),
		metav1.ListOptions{LabelSelector: utils.JoinSelector(appSelector)},
	)
	if err != nil {
		return nil, err
	}

	var count int
	var podSelector *metav1.LabelSelector
	for _, sts := range statefulsets.Items {
		if strings.Contains(sts.Name, "postgres") {
			f.dbms = api.Postgres
			f.statefulsetName = sts.Name
			podSelector = sts.Spec.Selector
			count++
		} else if strings.Contains(sts.Name, "mysql") {
			f.dbms = api.MySQL
			f.statefulsetName = sts.Name
			podSelector = sts.Spec.Selector
			count++
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("could not find a postgres or mysql statefulset for chart release %s", r.FullName())
	}
	if count != 1 {
		return nil, fmt.Errorf("expected exactly one mysql or postgres statefulset for chart release %s, found %d", r.FullName(), count)
	}

	if podSelector == nil {
		return nil, fmt.Errorf("statefulset %s has nil pod selector", f.statefulsetName)
	}

	// identify pod in statefulset
	pods, err := k8sclient.CoreV1().Pods(r.Namespace()).List(
		context.Background(),
		metav1.ListOptions{LabelSelector: utils.JoinSelector(podSelector.MatchLabels)},
	)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("could not find a running database pod in statefulset %s", f.statefulsetName)
	}

	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("expect exactly one running pod in statefulset %s, found %d", f.statefulsetName, len(pods.Items))
	}

	pod := pods.Items[0]
	if len(pod.Spec.Containers) != 1 {
		return nil, fmt.Errorf("expect exactly one container in pod %s, found %d", pod.Name, len(pod.Spec.Containers))
	}
	f.container = kubectl.Container{
		Pod:       pod.Name,
		Namespace: r.Namespace(),
		Name:      pod.Spec.Containers[0].Name,
	}

	// identify service
	services, err := k8sclient.CoreV1().Services(r.Namespace()).List(
		context.Background(),
		metav1.ListOptions{LabelSelector: utils.JoinSelector(appSelector)},
	)
	if err != nil {
		return nil, err
	}

	matchName := strings.ToLower(f.dbms.String())
	count = 0
	for _, s := range services.Items {
		if strings.Contains(s.Name, matchName) {
			count++
			f.serviceName = s.Name
		}
	}
	if count == 0 {
		return nil, fmt.Errorf("could not find service for statefulset %s", f.statefulsetName)
	}
	if count > 1 {
		return nil, fmt.Errorf("expect exactly one service for statefulset %s, found %d", f.statefulsetName, count)
	}

	f.serviceHostName = fmt.Sprintf("%s.%s.svc.cluster.local", f.serviceName, r.Namespace())
	return &f, nil
}
