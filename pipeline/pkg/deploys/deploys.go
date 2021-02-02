package deploys

import (
	"errors"
	"fmt"
	"os"
)

// Versions are, practically, numbers, but we use strings for
// potential extensibility.
type VersionType string

// "Legacy" versions are ones which predate the CDN domain and bucket.
const LegacyVersion VersionType = "LEGACY"

type DeployType string

const (
	DeployTesting    DeployType = "testing"
	DeployStaging    DeployType = "staging"
	DeployProduction DeployType = "prod"
	DeployUnknown    DeployType = ""
)

// Describes which deploys go where; in the legacy version, they're in
// the same bucket but separate directories; in the non-legacy
// version, they're in the top level of separate buckets, at separate
// domains names.
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

// Reads the DEPLOY environment variable; defaults to DeployTesting if unset.
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

// Fills out the bucket information from the TESTING_BUCKET
// environment variable if in testing.
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

// Returns the gs:// URL that files in the bucket can be uploaded to,
// for the given API version; never ends with a `/`.
func GetUploadURL(version VersionType) (string, error) {
	config, err := getDeployConfig()
	if err != nil {
		return "", err
	}
	return config.GetUploadURL(version), nil
}

// Returns the https:// URL that files in the bucket can be read from,
// for the given API version; never ends with a `/`.
func GetDownloadURL(version VersionType) (string, error) {
	config, err := getDeployConfig()
	if err != nil {
		return "", err
	}
	return config.GetDownloadURL(version), nil
}
