package bucket

import "path"

// CloudConsoleObjectDetailURL returns a URL displaying an object detail view in the Google cloud console.
// See https://cloud.google.com/storage/docs/request-endpoints#console
// Returns https://console.cloud.google.com/storage/browser/_details/<BUCKET_NAME>/<OBJECT_NAME>
func CloudConsoleObjectDetailURL(bucketName string, objectName string) string {
	return cloudConsoleBaseURL + "/" + path.Join("_details", bucketName, objectName)
}

// CloudConsoleObjectListURL returns a URL displaying a list of objects under a given path
// prefix in the Google cloud console.
// See https://cloud.google.com/storage/docs/request-endpoints#console
// Returns https://console.cloud.google.com/storage/browser/<BUCKET_NAME>/<PATH_PREFIX>
func CloudConsoleObjectListURL(bucketName string, pathPrefix string) string {
	return cloudConsoleBaseURL + "/" + path.Join(bucketName, pathPrefix)
}
