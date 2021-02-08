package main

import (
	"context"
	"flag"
	"os"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/metrics"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/secrets"
)

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	noopFlag := flag.Bool("noop", false, "Only print output, don't upload")
	metricsFlag := flag.Bool("metrics", false, "Enable metrics reporting")
	flag.Parse()

	secrets.RequireAirtableSecret()

	if *metricsFlag {
		metricsCleanup := metrics.Init()
		defer metricsCleanup()
	}

	var pm *generator.PublishManager
	if *noopFlag {
		pm = generator.NewNoopPublishManager()
	} else {
		pm = generator.NewPublishManager()
	}
	ok := pm.PublishAll(context.Background())
	if !ok {
		os.Exit(1)
	}
}
