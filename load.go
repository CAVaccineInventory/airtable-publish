package main

import (
	"github.com/pkg/errors"
	"log"
	"os/exec"
	"path"
)

// uploadFile uploads a file from disk to a Google Cloud Storage bucket.
func uploadFile(tableName string, destinationFile string) error {
	sourceFile := tableName + ".json"
	// TODO: consider doing this in Go directly. But last I recall, the Go SDK was a bit fussy with Go modules...
	cmd := exec.Command("gsutil", "-h", "Cache-Control:public,max-age=300", "cp", "-Z", path.Join(readyDir, sourceFile), destinationFile)
	if uploadErr := cmd.Run(); uploadErr != nil {
		log.Println(cmd.Output())
		return errors.Wrapf(uploadErr, "failed to upload json file %s to %s", sourceFile, destinationFile)
	}
	return nil
}
