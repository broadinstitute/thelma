package terra

type Cluster interface {
	// Address is the API address of the cluster master endpoint, eg. "https://35.36.37.38"
	Address() string
	// Project is the name of the GCP project where the cluster lives
	Project() string
	// Location is the name of the zone or region where the cluster lives (eg. "us-central1-a")
	Location() string
	Destination
}
