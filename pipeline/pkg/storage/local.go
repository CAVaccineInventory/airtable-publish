package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/metadata"
	beeline "github.com/honeycombio/beeline-go"
)

// StoreLocal writes out to a `local/` directory
func StoreLocal(ctx context.Context, destinationFile string, transformedData metadata.JSONData) error {
	ctx, span := beeline.StartSpan(ctx, "storage.StoreLocal")
	defer span.Send()
	beeline.AddField(ctx, "destinationFile", destinationFile)

	serializedData, err := Serialize(transformedData)
	if err != nil {
		err = fmt.Errorf("failed to serialize json: %w", err)
		beeline.AddField(ctx, "error", err)
		return err
	}

	writeToURL, err := url.Parse(destinationFile)
	if err != nil {
		err = fmt.Errorf("Invalid destination URL %s: %w", destinationFile, err)
		beeline.AddField(ctx, "error", err)
		return err
	}

	// Strip of the gs://; the "host" is the bucket name, which in
	// treat as a directory.  In most cases this is "local", which,
	// in Docker, is mounted out to the host OS.
	localFilePath := filepath.Join(writeToURL.Host, writeToURL.Path)
	err = os.MkdirAll(path.Dir(localFilePath), 0755)
	if err != nil {
		err = fmt.Errorf("failed to make directories %s: %w", path.Dir(localFilePath), err)
		beeline.AddField(ctx, "error", err)
		return err
	}

	err = ioutil.WriteFile(localFilePath, serializedData.Bytes(), 0644)
	log.Printf("Wrote out to local path: %s", localFilePath)
	if err != nil {
		err = fmt.Errorf("failed to write serialized json: %w", err)
		beeline.AddField(ctx, "error", err)
		return err
	}

	return nil
}
