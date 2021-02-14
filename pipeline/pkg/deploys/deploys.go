package deploys

import (
	"fmt"
	"os"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/metadata"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
)

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
var deploys = map[DeployType]*DeployConfig{
	DeployTesting: {
		// The bucket name here used for the name of the local
		// directory to write into.
		Storage: storage.StoreLocal,
		LegacyBucket: Bucket{
			Name: "local",
			Path: "legacy",
		},
		APIBucket: Bucket{
			Name: "local",
			Path: "api",
		},
	},
	DeployStaging: {
		Storage: storage.UploadToGCS,
		LegacyBucket: Bucket{
			Name: "cavaccineinventory-sitedata",
			Path: "airtable-sync-staging",
		},
		APIBucket: Bucket{
			Name:     "vaccinateca-api-staging",
			HostedAt: "staging-api.vaccinateca.com",
		},
	},
	DeployProduction: {
		Storage: storage.UploadToGCS,
		LegacyBucket: Bucket{
			Name: "cavaccineinventory-sitedata",
			Path: "airtable-sync",
		},
		APIBucket: Bucket{
			Name:     "vaccinateca-api",
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

// Returns the deployment configuration.
func getDeployConfig() (*DeployConfig, error) {
	deploy, err := GetDeploy()
	if err != nil {
		return nil, err
	}
	config := deploys[deploy]
	return config, nil
}

func GetStorage() (StorageWriter, error) {
	config, err := getDeployConfig()
	if err != nil {
		return nil, err
	}
	return config.Storage, nil
}

func SetTestingStorage(sw StorageWriter, bucketName string) {
	deploys[DeployTesting].Storage = sw
	deploys[DeployTesting].LegacyBucket.Name = bucketName
	deploys[DeployTesting].APIBucket.Name = bucketName
}

// Returns the gs:// URL that files in the bucket can be uploaded to,
// for the given API version; never ends with a `/`.
func GetUploadURL(version metadata.VersionType) (string, error) {
	config, err := getDeployConfig()
	if err != nil {
		return "", err
	}
	return config.GetUploadURL(version), nil
}

// Returns the https:// URL that files in the bucket can be read from,
// for the given API version; never ends with a `/`.
func GetDownloadURL(version metadata.VersionType) (string, error) {
	config, err := getDeployConfig()
	if err != nil {
		return "", err
	}
	return config.GetDownloadURL(version), nil
}
