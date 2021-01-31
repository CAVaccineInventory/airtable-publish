package storage

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	beeline "github.com/honeycombio/beeline-go"
)

func UploadToGCS(ctx context.Context, destinationFile string, transformedData airtable.TableContent) error {
	ctx, span := beeline.StartSpan(ctx, "storage.UploadToGCS")
	defer span.Send()
	beeline.AddField(ctx, "destinationFile", destinationFile)

	tempDir, err := ioutil.TempDir("", "gcs-upload")
	defer os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("failed to make temp directory: %w", err)
	}
	localFile := path.Join(tempDir, "output.json")
	f, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localFile, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	serializedData, err := transformedData.Serialize()
	if err != nil {
		return fmt.Errorf("failed to write serialized json: %w", err)
	}
	_, err = w.Write(serializedData.Bytes())
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
