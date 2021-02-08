package main

import (
	"context"
	"flag"
	"os"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
)

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	noopFlag := flag.Bool("noop", false, "Only print output, don't upload")
	flag.Parse()

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
