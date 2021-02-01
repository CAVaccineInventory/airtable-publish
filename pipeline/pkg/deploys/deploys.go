package deploys

import (
	"errors"
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
	DeployTesting:    "", // Must be set by TESTING_BUCKET env var
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

func getPath() (string, error) {
	deploy, err := GetDeploy()
	if err != nil {
		return "", err
	}

	if deploy != DeployTesting {
		return exportPaths[deploy], nil
	}

	path := os.Getenv("TESTING_BUCKET")
	if path == "" {
		return "", errors.New("Set TESTING_BUCKET env var to the name of your bucket (see README.md)")
	}
	return path, nil
}

func GetUploadURL() (string, error) {
	path, err := getPath()
	if err != nil {
		return "", err
	}
	return "gs://" + path, nil
}

func GetDownloadURL() (string, error) {
	path, err := getPath()
	if err != nil {
		return "", err
	}
	return exportBaseURL + path, nil
}
