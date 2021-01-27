package main

import (
	"log"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	totalPublishLatency = stats.Float64(
		"total_publish_latency_s",
		"Latency to extract and publish all measures",
		stats.UnitSeconds,
	)

	tablePublishLatency = stats.Float64(
		"table_publish_latency_s",
		"Latency to extract and publish each table",
		stats.UnitSeconds,
	)

	tablePublishFailures = stats.Int64(
		"table_publish_failures_count",
		"Number of failed publishes by table",
		stats.UnitDimensionless,
	)

	keyTable, _  = tag.NewKey("table")
	keyDeploy, _ = tag.NewKey("deploy")
)

func InitMetrics() func() {
	err := view.Register(
		&view.View{
			Name:        totalPublishLatency.Name(),
			Description: totalPublishLatency.Description(),
			Measure:     totalPublishLatency,
			// These are large latency buckets because
			// fetches from Airtable take multiple seconds
			Aggregation: view.Distribution(
				30, 40, 50, 60, 75, 100, 120, 150, 185, 230, 290,
			),
			TagKeys: []tag.Key{keyDeploy},
		},
		&view.View{
			Name:        tablePublishLatency.Name(),
			Description: tablePublishLatency.Description(),
			Measure:     tablePublishLatency,
			Aggregation: view.Distribution(
				7, 10, 13, 20, 26, 37, 73, 100, 145, 200,
			),
			TagKeys: []tag.Key{keyDeploy, keyTable},
		},
		&view.View{
			Name:        "table_publish_count",
			Description: "Total number of publishes by table",
			Measure:     tablePublishLatency,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{keyDeploy, keyTable},
		},
		&view.View{
			Name:        tablePublishFailures.Name(),
			Description: tablePublishFailures.Description(),
			Measure:     tablePublishFailures,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{keyDeploy, keyTable},
		},
	)
	if err != nil {
		log.Fatalf("Failed to register the view: %v", err)
	}

	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID:         "cavaccineinventory",
		ReportingInterval: 60 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := exporter.StartMetricsExporter(); err != nil {
		log.Fatalf("Error starting metric exporter: %v", err)
	}

	return func() {
		exporter.Flush()
		exporter.StopMetricsExporter()
	}
}
