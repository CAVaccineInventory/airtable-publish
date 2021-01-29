package main

import (
	"log"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	beeline "github.com/honeycombio/beeline-go"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	keyDeploy, _ = tag.NewKey("deploy")
)

func InitMetrics() func() {
	beeline.Init(beeline.Config{
		WriteKey:    os.Getenv("HONEYCOMB_KEY"),
		Dataset:     "pipeline",
		ServiceName: "pipeline",
	})
	err := view.Register(
		&view.View{
			Name:        generator.TotalPublishLatency.Name(),
			Description: generator.TotalPublishLatency.Description(),
			Measure:     generator.TotalPublishLatency,
			// These are large latency buckets because
			// fetches from Airtable take multiple seconds
			Aggregation: view.Distribution(
				30, 40, 50, 60, 75, 100, 120, 150, 185, 230, 290,
			),
			TagKeys: []tag.Key{keyDeploy},
		},
		&view.View{
			Name:        generator.TablePublishLatency.Name(),
			Description: generator.TablePublishLatency.Description(),
			Measure:     generator.TablePublishLatency,
			Aggregation: view.Distribution(
				7, 10, 13, 20, 26, 37, 73, 100, 145, 200,
			),
			TagKeys: []tag.Key{keyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        airtable.FetchLatency.Name(),
			Description: airtable.FetchLatency.Description(),
			Measure:     airtable.FetchLatency,
			Aggregation: view.Distribution(
				7, 10, 13, 20, 26, 37, 73, 100, 145, 200,
			),
			TagKeys: []tag.Key{keyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        "table_publish_count",
			Description: "Total number of publishes by table",
			Measure:     generator.TablePublishLatency,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{keyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        generator.TablePublishSuccesses.Name(),
			Description: generator.TablePublishSuccesses.Description(),
			Measure:     generator.TablePublishSuccesses,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{keyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        generator.TablePublishFailures.Name(),
			Description: generator.TablePublishFailures.Description(),
			Measure:     generator.TablePublishFailures,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{keyDeploy, generator.KeyTable},
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
		beeline.Close()

		exporter.Flush()
		exporter.StopMetricsExporter()
	}
}
