// Package main contains a Functions Framework wrapper.
package main

import (
	"net/http"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/handlers"
	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/metrics"
)

func main() {
	metricsCleanup := metrics.Init()
	defer metricsCleanup()

	// Serve health status.
	http.HandleFunc("/", handlers.Health)
	http.HandleFunc("/json", handlers.ExportJSON)
	http.HandleFunc("/push", handlers.PushMetrics)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
