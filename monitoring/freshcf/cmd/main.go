// Package main contains a Functions Framework wrapper.
package main

import (
	"context"
	"log"
	"os"

	"github.com/CAVaccineInventory/airtable-export/freshcf"
	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	ctx := context.Background()
	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/", freshcf.CheckFreshness); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext /: %v\n", err)
	}

	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/json", freshcf.ExportJSON); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext /json: %v\n", err)
	}

	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
