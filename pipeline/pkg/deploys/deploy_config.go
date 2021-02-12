package deploys

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/metadata"
)

// A function which can be used to output the transformed data; see pkg/storage/
type StorageWriter func(ctx context.Context, destinationFile string, transformedData metadata.JSONData) error

type Bucket struct {
	Name     string
	Path     string
	HostedAt string
}

// Returns the gs:// URL that files in the bucket can be uploaded to;
// never ends with a `/`.
func (b *Bucket) GetUploadURL() string {
	path := ""
	if b.Path != "" {
		path = "/" + b.Path
	}
	return "gs://" + b.Name + path
}

// Returns the https:// URL that files in the bucket can be read from;
// never ends with a `/`.
func (b *Bucket) GetDownloadURL() string {
	path := ""
	if b.Path != "" {
		path = "/" + b.Path
	}
	if b.HostedAt == "" {
		return "https://storage.googleapis.com/" + b.Name + path
	}
	return "https://" + b.HostedAt + path
}

// Pair of legacy bucket, and new bucket hooked up to a CDN
type DeployConfig struct {
	Storage      StorageWriter
	LegacyBucket Bucket
	APIBucket    Bucket
}

// Returns the gs:// URL that files in the bucket can be uploaded to,
// for the given API version; never ends with a `/`.  Prefer to use
// deploys.GetUploadURL, which chooses the right DeployConfig based on
// the deploy environment.
func (dc *DeployConfig) GetUploadURL(version VersionType) string {
	if version == LegacyVersion {
		return dc.LegacyBucket.GetUploadURL()
	}
	return fmt.Sprintf("%s/v%s", dc.APIBucket.GetUploadURL(), version)
}

// Returns the https:// URL that files in the bucket can be read from,
// for the given API version; never ends with a `/`.  Prefer to use
// deploys.GetDownloadURL, which chooses the right DeployConfig based on
// the deploy environment.
func (dc *DeployConfig) GetDownloadURL(version VersionType) string {
	if version == LegacyVersion {
		return dc.LegacyBucket.GetDownloadURL()
	}
	return fmt.Sprintf("%s/v%s", dc.APIBucket.GetDownloadURL(), version)
}
