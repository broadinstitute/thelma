package terra

type Cluster interface {
	// Address is the API address of the cluster master endpoint, eg. "https://35.36.37.38"
	Address() string
	// Project is the name of the GCP project where the cluster lives
	Project() string
	// ProjectSuffix is the last hyphen-delimited part of the GCP project name
	ProjectSuffix() string
	// Location is the name of the zone or region where the cluster lives (eg. "us-central1-a")
	Location() string
	// ArtifactBucket is the name of a GCS bucket for the cluster where Thelma can upload artifacts
	ArtifactBucket() string
	Destination
}
