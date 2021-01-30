package freshcf

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/deploys"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	lastModified = stats.Int64(
		"last_modified_s",
		"Raw last-modified time, parsed from the header, in seconds since epoch",
		stats.UnitSeconds,
	)

	lastModifiedAge = stats.Float64(
		"last_modified_age_s",
		"Latency to fetched Last-Modified header",
		stats.UnitSeconds,
	)

	fileLengthBytes = stats.Int64(
		"file_length_bytes",
		"Length of the fetched document",
		stats.UnitBytes,
	)

	fileLengthJSONItems = stats.Int64(
		"file_length_json_items",
		"How many items are in the list at the top level of the fetched JSON document",
		stats.UnitDimensionless,
	)

	keyTable, _  = tag.NewKey("table")
	keyDeploy, _ = tag.NewKey("deploy")
)

var tableNames = [...]string{"Locations", "Counties"}

func InitMetrics() func() {
	err := view.Register(
		&view.View{
			Name:        lastModified.Name(),
			Description: lastModified.Description(),
			Measure:     lastModified,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{keyDeploy, keyTable},
		},
		&view.View{
			Name:        lastModifiedAge.Name(),
			Description: lastModifiedAge.Description(),
			Measure:     lastModifiedAge,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{keyDeploy, keyTable},
		},
		&view.View{
			Name:        fileLengthBytes.Name(),
			Description: fileLengthBytes.Description(),
			Measure:     fileLengthBytes,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{keyDeploy, keyTable},
		},
		&view.View{
			Name:        fileLengthJSONItems.Name(),
			Description: fileLengthJSONItems.Description(),
			Measure:     fileLengthJSONItems,
			Aggregation: view.LastValue(),
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

func PushMetrics(w http.ResponseWriter, r *http.Request) {
	deploy, err := deploys.GetDeploy()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error determining deploy: %v", err)
		return
	}
	deployCtx, _ := tag.New(context.Background(), tag.Insert(keyDeploy, string(deploy)))

	baseURL, err := deploys.GetExportBaseURL()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error determining base url: %v", err)
		return
	}

	for _, tableName := range tableNames {
		tableCtx, _ := tag.New(deployCtx, tag.Insert(keyTable, tableName))

		url := fmt.Sprintf("%v/%v.json", baseURL, tableName)
		urlStats, err := getURLStats(url)

		if err != nil {
			log.Printf("error getting %q stats %q: %v", tableName, url, err)
			// XXX FUTURE report a count of errors
		}

		stats.Record(tableCtx, fileLengthBytes.M(int64(urlStats.FileLengthBytes)))
		stats.Record(tableCtx, fileLengthJSONItems.M(int64(urlStats.FileLengthJSONItems)))

		if !urlStats.LastModified.IsZero() {
			stats.Record(tableCtx, lastModified.M(urlStats.LastModified.Unix()))
			ago := time.Since(urlStats.LastModified).Seconds()
			stats.Record(tableCtx, lastModifiedAge.M(ago))
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "done")
}
