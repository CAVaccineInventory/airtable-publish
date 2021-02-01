package deploys

import "fmt"

type Bucket struct {
	Name     string
	Path     string
	HostedAt string
}

func (b *Bucket) GetUploadURL() string {
	path := ""
	if b.Path != "" {
		path = "/" + b.Path
	}
	return "gs://" + b.Name + path
}
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

type DeployConfig struct {
	LegacyBucket Bucket
	APIBucket    Bucket
}

func (dc *DeployConfig) GetUploadURL(version VersionType) string {
	if version == LegacyVersion {
		return dc.LegacyBucket.GetUploadURL()
	}
	return fmt.Sprintf("%s/v%s", dc.APIBucket.GetUploadURL(), version)
}
func (dc *DeployConfig) GetDownloadURL(version VersionType) string {
	if version == LegacyVersion {
		return dc.LegacyBucket.GetDownloadURL()
	}
	return fmt.Sprintf("%s/v%s", dc.APIBucket.GetDownloadURL(), version)
}
