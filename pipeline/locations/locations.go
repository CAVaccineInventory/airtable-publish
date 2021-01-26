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
	DeployTesting:    "cavaccineinventory-sitedata/airtable-sync-testing",
	DeployStaging:    "cavaccineinventory-sitedata/airtable-sync-staging",
	DeployProduction: "cavaccineinventory-sitedata/airtable-sync",
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
