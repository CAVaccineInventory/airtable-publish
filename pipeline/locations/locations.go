package locations

import (
	"fmt"
	"os"
)

type DeployType string

const (
	DeployTesting    DeployType = "testing"
	DeployStaging    DeployType = "staging"
	DeployProduction DeployType = "prod"
	DeployUnknown    DeployType = ""
)

var exportPaths = map[DeployType]string{
	DeployTesting:    "cavaccineinventory-sitedata/airtable-sync-testing",
	DeployStaging:    "cavaccineinventory-sitedata/airtable-sync-staging",
	DeployProduction: "cavaccineinventory-sitedata/airtable-sync",
}

const exportBaseURL = "https://storage.googleapis.com/"

func GetDeploy() (DeployType, error) {
	deploy := DeployType(os.Getenv("DEPLOY"))
	if deploy == DeployUnknown {
		deploy = DeployTesting
	}
	if _, ok := exportPaths[deploy]; !ok {
		return DeployUnknown, fmt.Errorf("Unknown deploy environment: %s", deploy)
	}
	return deploy, nil
}

func GetExportBucket() (string, error) {
	deploy, err := GetDeploy()
	if err != nil {
		return "", err
	}
	return "gs://" + exportPaths[deploy], nil
}

func GetExportBaseURL() (string, error) {
	deploy, err := GetDeploy()
	if err != nil {
		return "", err
	}
	return exportBaseURL + exportPaths[deploy], nil
}
