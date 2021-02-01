package deploys

import (
	"errors"
	"fmt"
	"os"
)

type VersionType string

const LegacyVersion VersionType = "LEGACY"

type DeployType string

const (
	DeployTesting    DeployType = "testing"
	DeployStaging    DeployType = "staging"
	DeployProduction DeployType = "prod"
	DeployUnknown    DeployType = ""
)

var deploys = map[DeployType]DeployConfig{
	DeployTesting: {
		LegacyBucket: Bucket{
			// name is set below
			Path: "legacy",
		},
		APIBucket: Bucket{
			// name is set below
			Path: "api",
		},
	},
	DeployStaging: {
		LegacyBucket: Bucket{
			Name: "cavaccineinventory-sitedata",
			Path: "airtable-sync-staging",
		},
		APIBucket: Bucket{
			Name:     "vaccinataca-api-staging",
			HostedAt: "staging-api.vaccinateca.com",
		},
	},
	DeployProduction: {
		LegacyBucket: Bucket{
			Name: "cavaccineinventory-sitedata",
			Path: "airtable-sync",
		},
		APIBucket: Bucket{
			Name:     "vaccinataca-api",
			HostedAt: "api.vaccinateca.com",
		},
	},
}

func GetDeploy() (DeployType, error) {
	deploy := DeployType(os.Getenv("DEPLOY"))
	if deploy == DeployUnknown {
		deploy = DeployTesting
	}
	if _, ok := deploys[deploy]; !ok {
		return DeployUnknown, fmt.Errorf("Unknown deploy environment: %s", deploy)
	}
	return deploy, nil
}

func getDeployConfig() (*DeployConfig, error) {
	deploy, err := GetDeploy()
	if err != nil {
		return nil, err
	}
	config := deploys[deploy]
	if deploy != DeployTesting {
		return &config, nil
	}

	bucketName := os.Getenv("TESTING_BUCKET")
	if bucketName == "" {
		return nil, errors.New("Set TESTING_BUCKET env var to the name of your bucket (see README.md)")
	}
	config.LegacyBucket.Name = bucketName
	config.APIBucket.Name = bucketName
	return &config, nil
}

func GetUploadURL(version VersionType) (string, error) {
	config, err := getDeployConfig()
	if err != nil {
		return "", err
	}
	return config.GetUploadURL(version), nil
}

func GetDownloadURL(version VersionType) (string, error) {
	config, err := getDeployConfig()
	if err != nil {
		return "", err
	}
	return config.GetDownloadURL(version), nil
}
