package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/metrics"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/secrets"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
)

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	noopFlag := flag.Bool("noop", false, "Only print output, don't upload")
	bucketFlag := flag.String("bucket", "", "Upload into a specific bucket")
	metricsFlag := flag.Bool("metrics", false, "Enable metrics reporting")
	flag.Parse()

	if *noopFlag && *bucketFlag != "" {
		log.Fatal("-noop and -bucket are mutually exclusive!")
	}

	ctx := context.Background()
	sec := secrets.RequireAirtableSecret(ctx)

	if *metricsFlag {
		ctx, cxl := context.WithTimeout(ctx, 30*time.Second)
		defer cxl()
		metricsCleanup := metrics.Init(ctx)
		defer metricsCleanup()
	}

	pm := generator.NewPublishManager()

	if *noopFlag {
		// "bucket-name" is arbitrary here, since nothing is written anywhere
		deploys.SetTestingStorage(storage.DebugToSTDERR, "bucket-name")
	} else if *bucketFlag != "" {
		deploys.SetTestingStorage(storage.UploadToGCS, *bucketFlag)
	}

	tables := airtable.NewTables(sec)
	ok := pm.PublishAll(ctx, tables)
	if !ok {
		os.Exit(1)
	}
}
