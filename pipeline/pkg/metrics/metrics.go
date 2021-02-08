package metrics

import (
	"context"
	"fmt"
	"log"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/config"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/secrets"
	beeline "github.com/honeycombio/beeline-go"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	KeyDeploy, _ = tag.NewKey("deploy")
)

// Set up Honeycomb and Stackdriver (via OpenCensus) metric logging.
// Returns a cleanup function which should be called before exit, to
// push any final metrics.
func Init() func() {
	deploy, err := deploys.GetDeploy()
	if err != nil {
		log.Fatal(err)
	}
	honeycombKey, err := secrets.Get(context.Background(), secrets.HoneycombSecret)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed to get Honeycomb credentials: %w", err))
	}
	beeline.Init(beeline.Config{
		WriteKey:    honeycombKey,
		Dataset:     fmt.Sprintf("pipeline-%s", deploy),
		ServiceName: "pipeline",
	})
	err = view.Register(
		&view.View{
			Name:        generator.TotalPublishLatency.Name(),
			Description: generator.TotalPublishLatency.Description(),
			Measure:     generator.TotalPublishLatency,
			// These are large latency buckets because
			// fetches from Airtable take multiple seconds
			Aggregation: view.Distribution(
				30, 40, 50, 60, 75, 100, 120, 150, 185, 230, 290,
			),
			TagKeys: []tag.Key{KeyDeploy},
		},
		&view.View{
			Name:        generator.TablePublishLatency.Name(),
			Description: generator.TablePublishLatency.Description(),
			Measure:     generator.TablePublishLatency,
			Aggregation: view.Distribution(
				7, 10, 13, 20, 26, 37, 73, 100, 145, 200,
			),
			TagKeys: []tag.Key{KeyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        airtable.FetchLatency.Name(),
			Description: airtable.FetchLatency.Description(),
			Measure:     airtable.FetchLatency,
			Aggregation: view.Distribution(
				7, 10, 13, 20, 26, 37, 73, 100, 145, 200,
			),
			TagKeys: []tag.Key{KeyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        "table_publish_count",
			Description: "Total number of publishes by table",
			Measure:     generator.TablePublishLatency,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{KeyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        generator.TablePublishSuccesses.Name(),
			Description: generator.TablePublishSuccesses.Description(),
			Measure:     generator.TablePublishSuccesses,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{KeyDeploy, generator.KeyTable},
		},
		&view.View{
			Name:        generator.TablePublishFailures.Name(),
			Description: generator.TablePublishFailures.Description(),
			Measure:     generator.TablePublishFailures,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{KeyDeploy, generator.KeyTable},
		},
	)
	if err != nil {
		log.Fatalf("Failed to register the view: %v", err)
	}

	exporter, err := stackdriver.NewExporter(config.StackdriverOptions("pipeline"))
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
