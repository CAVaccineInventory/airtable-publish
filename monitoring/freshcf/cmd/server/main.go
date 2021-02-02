// Package main contains a Functions Framework wrapper.
package main

import (
	"net/http"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf"
	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/metrics"
)

func main() {
	metricsCleanup := metrics.Init()
	defer metricsCleanup()

	// Serve health status.
	http.HandleFunc("/", freshcf.CheckFreshness)
	http.HandleFunc("/json", freshcf.ExportJSON)
	http.HandleFunc("/push", freshcf.PushMetrics)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
