package main

import (
	"context"
	"flag"
	"os"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
)

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	debugFlag := flag.Bool("debug", false, "Only print output, don't upload")
	flag.Parse()

	var pm *generator.PublishManager
	if *debugFlag {
		pm = generator.NewDebugPublishManager()
	} else {
		pm = generator.NewPublishManager()
	}
	ok := pm.PublishAll(context.Background(), endpoints.AllEndpoints)
	if !ok {
		os.Exit(1)
	}
}
