package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

// uploadFile uploads a file from disk to a Google Cloud Storage bucket.
func uploadFile(ctx context.Context, sourceFile string, destinationFile string) error {
	// TODO: consider doing this in Go directly. But last I recall, the Go SDK was a bit fussy with Go modules...

	// Update the README.md for new latencies if you adjust the max-age.
	cmd := exec.CommandContext(ctx, "gsutil", "-h", "Cache-Control:public,max-age=120", "cp", "-Z", sourceFile, destinationFile)
	output, uploadErr := cmd.CombinedOutput()
	if uploadErr != nil {
		log.Println(string(output))
		return fmt.Errorf("failed to upload json file %s to %s: %w", sourceFile, destinationFile, uploadErr)
	}
	return nil
}
