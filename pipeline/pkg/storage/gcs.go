package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	beeline "github.com/honeycombio/beeline-go"
)

func UploadToGCS(ctx context.Context, destinationFile string, transformedData airtable.TableContent) error {
	ctx, span := beeline.StartSpan(ctx, "storage.UploadToGCS")
	defer span.Send()
	beeline.AddField(ctx, "destinationFile", destinationFile)

	serializedData, err := transformedData.Serialize()
	if err != nil {
		return fmt.Errorf("failed to write serialized json: %w", err)
	}

	tempDir, err := ioutil.TempDir("", "gcs-upload")
	defer os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("failed to make temp directory: %w", err)
	}
	localFile := filepath.Join(tempDir, "output.json")

	err = ioutil.WriteFile(localFile, serializedData.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("failed to write sanitized json to %s: %w", localFile, err)
	}

	// TODO: consider doing this in Go directly. But last I recall, the Go SDK was a bit fussy with Go modules...

	// Update the README.md for new latencies if you adjust the max-age.
	cmd := exec.CommandContext(ctx, "gsutil", "-h", "Cache-Control:public,max-age=120", "cp", "-Z", localFile, destinationFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(output))
		return fmt.Errorf("failed to upload json file %s to %s: %w", localFile, destinationFile, err)
	}
	return nil
}
