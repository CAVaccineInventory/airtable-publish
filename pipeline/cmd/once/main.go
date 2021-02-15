package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

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

	secrets.RequireAirtableSecret()

	if *metricsFlag {
		ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
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

	ok := pm.PublishAll(context.Background())
	if !ok {
		os.Exit(1)
	}
}
