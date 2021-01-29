package storage

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	beeline "github.com/honeycombio/beeline-go"
)

func UploadToGCS(ctx context.Context, tableName string, sourceFile string) error {
	ctx, span := beeline.StartSpan(ctx, "storage.UploadToGCS")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	bucket, err := deploys.GetExportBucket()
	if err != nil {
		return fmt.Errorf("[%s] failed to get destination bucket: %w", tableName, err)
	}
	destinationFile := bucket + "/" + tableName + ".json"
	// TODO: consider doing this in Go directly. But last I recall, the Go SDK was a bit fussy with Go modules...

	// Update the README.md for new latencies if you adjust the max-age.
	cmd := exec.CommandContext(ctx, "gsutil", "-h", "Cache-Control:public,max-age=120", "cp", "-Z", sourceFile, destinationFile)
	output, uploadErr := cmd.CombinedOutput()
	if uploadErr != nil {
		log.Println(string(output))
		return fmt.Errorf("[%s] failed to upload json file %s to %s: %w", tableName, sourceFile, destinationFile, uploadErr)
	}
	return nil
}
