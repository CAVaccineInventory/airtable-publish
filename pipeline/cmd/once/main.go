package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/metrics"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/secrets"
)

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	noopFlag := flag.Bool("noop", false, "Only print output, don't upload")
	localFlag := flag.Bool("local", false, "Write to files under local/")
	metricsFlag := flag.Bool("metrics", false, "Enable metrics reporting")
	flag.Parse()

	if *localFlag && *noopFlag {
		log.Fatal("-noop and -local are mutually exclusive!")
	}

	secrets.RequireAirtableSecret()

	if *metricsFlag {
		metricsCleanup := metrics.Init()
		defer metricsCleanup()
	}

	var pm *generator.PublishManager
	if *noopFlag {
		pm = generator.NewNoopPublishManager()
	} else if *localFlag {
		pm = generator.NewLocalPublishManager()
	} else {
		pm = generator.NewPublishManager()
	}
	ok := pm.PublishAll(context.Background())
	if !ok {
		os.Exit(1)
	}
}
