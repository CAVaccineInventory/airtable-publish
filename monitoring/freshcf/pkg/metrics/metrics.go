package metrics

import (
	"log"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	LastModified = stats.Int64(
		"last_modified_s",
		"Raw last-modified time, parsed from the header, in seconds since epoch",
		stats.UnitSeconds,
	)

	LastModifiedAge = stats.Float64(
		"last_modified_age_s",
		"Latency to fetched Last-Modified header",
		stats.UnitSeconds,
	)

	FileLengthBytes = stats.Int64(
		"file_length_bytes",
		"Length of the fetched document",
		stats.UnitBytes,
	)

	FileLengthJSONItems = stats.Int64(
		"file_length_json_items",
		"How many items are in the list at the top level of the fetched JSON document",
		stats.UnitDimensionless,
	)

	KeyDeploy, _   = tag.NewKey("deploy")
	KeyVersion, _  = tag.NewKey("version")
	KeyResource, _ = tag.NewKey("resource")
)

func Init() func() {
	err := view.Register(
		&view.View{
			Name:        LastModified.Name(),
			Description: LastModified.Description(),
			Measure:     LastModified,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{KeyDeploy, KeyVersion, KeyResource},
		},
		&view.View{
			Name:        LastModifiedAge.Name(),
			Description: LastModifiedAge.Description(),
			Measure:     LastModifiedAge,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{KeyDeploy, KeyVersion, KeyResource},
		},
		&view.View{
			Name:        FileLengthBytes.Name(),
			Description: FileLengthBytes.Description(),
			Measure:     FileLengthBytes,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{KeyDeploy, KeyVersion, KeyResource},
		},
		&view.View{
			Name:        FileLengthJSONItems.Name(),
			Description: FileLengthJSONItems.Description(),
			Measure:     FileLengthJSONItems,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{KeyDeploy, KeyVersion, KeyResource},
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
