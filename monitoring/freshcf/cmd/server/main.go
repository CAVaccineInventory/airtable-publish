// Package main contains a Functions Framework wrapper.
package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/handlers"
	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/metrics"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
	"github.com/honeycombio/beeline-go/wrappers/hnynethttp"
)

func main() {
	bucketFlag := flag.String("bucket", "", "Upload into a specific bucket")
	metricsFlag := flag.Bool("metrics", true, "Enable metrics reporting")
	flag.Parse()

	if *metricsFlag {
		ctx, cxl := context.WithTimeout(context.Background(), 30*time.Second)
		defer cxl()
		metricsCleanup := metrics.Init(ctx)
		defer metricsCleanup()
	}

	if *bucketFlag != "" {
		deploys.SetTestingStorage(storage.UploadToGCS, *bucketFlag)
	}

	// Serve health status.
	http.HandleFunc("/", handlers.Health)
	http.HandleFunc("/json", handlers.ExportJSON)
	http.HandleFunc("/push", handlers.PushMetrics)
	err := http.ListenAndServe(":8080", hnynethttp.WrapHandler(http.DefaultServeMux))
	if err != nil {
		panic(err)
	}
}
