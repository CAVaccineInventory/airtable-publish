package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/metrics"
	freshstats "github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/stats"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func PushMetrics(w http.ResponseWriter, r *http.Request) {
	deploy, err := deploys.GetDeploy()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error determining deploy: %v", err)
		return
	}
	ctx, _ := tag.New(r.Context(), tag.Insert(metrics.KeyDeploy, string(deploy)))

	resultChan := freshstats.AllResponses(ctx)
	for len(resultChan) != 0 {
		result := <-resultChan
		ep := result.Endpoint
		tableCtx, _ := tag.New(ctx,
			tag.Insert(metrics.KeyVersion, string(ep.Version)),
			tag.Insert(metrics.KeyResource, ep.Resource))
		if result.Err != nil {
			log.Printf("error getting %v stats: %v", &ep, result.Err)
			continue
			// XXX FUTURE report a count of errors
		}

		urlStats := result.Stats
		stats.Record(tableCtx, metrics.FileLengthBytes.M(int64(urlStats.FileLengthBytes)))
		stats.Record(tableCtx, metrics.FileLengthJSONItems.M(int64(urlStats.FileLengthJSONItems)))

		if !urlStats.LastModified.IsZero() {
			stats.Record(tableCtx, metrics.LastModified.M(urlStats.LastModified.Unix()))
			ago := time.Since(urlStats.LastModified).Seconds()
			stats.Record(tableCtx, metrics.LastModifiedAge.M(ago))
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "done")
}
