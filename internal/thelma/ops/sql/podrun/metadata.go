package podrun

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/ops/sql/api"
	metadecoder "github.com/broadinstitute/thelma/internal/thelma/ops/sql/podrun/meta"
	"github.com/rs/zerolog/log"
	gomaps "golang.org/x/exp/maps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/user"
	"time"
)

// prefix for labels and annotations used by Thelma
const labelPrefix = "thelma.terra.bio/"

// annotations that are added to each pod
type annotations struct {
	// DeleteAfter used by Thelma to determine when to delete old/orphaned pods
	DeleteAfter time.Time `key:"delete-after"`

	// Remaining annotations are purely informational

	// CreatedFor name of the user this pod was created for
	CreatedFor string `key:"created-for"`
	// Provider - api.Provider value
	Provider string `key:"provider"`
	// DBMS - api.DBMS value
	DBMS string `key:"dbms"`
	// Instance api.Instance the pod is connecting to
	Instance api.Instance `key:"instance"`
}

type labels struct {
	Owned bool `key:"owned"` // used by Thelma to identify pods that it owns
}

// additional standard kubernetes labels added to the pod
var standardLabels = map[string]string{
	"app.kubernetes.io/name":       "thelma-workload",
	"app.kubernetes.io/managed-by": "thelma",
}

type commonMetadata struct {
	labels      map[string]string
	annotations map[string]string
}

func (c *commonMetadata) write(obj *metav1.ObjectMeta) {
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	gomaps.Copy(obj.Labels, standardLabels)
	gomaps.Copy(obj.Labels, c.labels)
	gomaps.Copy(obj.Annotations, c.annotations)
}

func createMetadata(conn api.Connection) (*commonMetadata, error) {
	_labels, err := createLabels()
	if err != nil {
		return nil, fmt.Errorf("error encoding labels: %v", err)
	}
	_annotations, err := createAnnotations(conn)
	if err != nil {
		return nil, fmt.Errorf("error encoding annotations: %v", err)
	}

	return &commonMetadata{
		labels:      _labels,
		annotations: _annotations,
	}, nil
}

func createAnnotations(conn api.Connection) (map[string]string, error) {
	var a annotations
	// TODO we should really use @broadinstitute.org email (can pull from Sherlock or Google OAuth token)
	var username string
	u, err := user.Current()
	if err != nil {
		log.Warn().Err(err).Msgf("Could not identity local user")
	} else {
		username = u.Username
	}

	a.CreatedFor = username
	a.DeleteAfter = time.Now().Add(maxPodLifetimeSeconds * time.Second)
	a.Provider = conn.Provider.String()
	a.DBMS = conn.Provider.String()
	a.Instance = conn.Instance()

	return newDecoder[annotations]().ToMap(a)
}

func createLabels() (map[string]string, error) {
	var l labels
	l.Owned = true

	return newDecoder[labels]().ToMap(l)
}

func newDecoder[T any]() metadecoder.Decoder[T] {
	return metadecoder.Decoder[T]{
		Prefix:  labelPrefix,
		TagName: "key",
	}
}
