package main

import (
	"context"
	"os"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
)

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	pm := generator.NewPublishManager()
	ok := pm.PublishAll(context.Background(), endpoints.AllEndpoints)
	if !ok {
		os.Exit(1)
	}
}
