package locations

import (
	"os"
)

const (
	DeployTesting    = "testing"
	DeployStaging    = "staging"
	DeployProduction = "prod"
)

var exportBuckets = map[string]string{
	DeployTesting:    "gs://cavaccineinventory-sitedata/airtable-sync-testing",
	DeployStaging:    "gs://cavaccineinventory-sitedata/airtable-sync-staging",
	DeployProduction: "gs://cavaccineinventory-sitedata/airtable-sync",
}

const exportBaseURL = "https://storage.googleapis.com/"

func GetDeploy() string {
	deploy := os.Getenv("DEPLOY")
	if deploy == "" {
		deploy = DeployTesting
	}
	return deploy
}

func GetExportBucket() string {
	return exportBuckets[GetDeploy()]
}

func GetExportBaseURL() string {
	return exportBaseURL + GetExportBucket()
}
