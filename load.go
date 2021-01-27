package main

import (
	"github.com/pkg/errors"
	"log"
	"os/exec"
	"path"
)

// uploadFile uploads a file from disk to a Google Cloud Storage bucket.
func uploadFile(sourceFilePath string, destinationPath string) error {
	// TODO: consider doing this in Go directly. But last I recall, the Go SDK was a bit fussy with Go modules...
	cmd := exec.Command("gsutil", "-h", "Cache-Control:public,max-age=300", "cp", "-Z", path.Join(sourceFilePath), destinationPath)
	output, uploadErr := cmd.CombinedOutput()
	if uploadErr != nil {
		log.Println(string(output))
		return errors.Wrapf(uploadErr, "failed to upload json file %s to %s", sourceFilePath, destinationPath)
	}
	return nil
}
