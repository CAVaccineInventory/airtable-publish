package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/metrics"
	freshstats "github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/stats"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints"
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
	deployCtx, _ := tag.New(context.Background(), tag.Insert(metrics.KeyDeploy, string(deploy)))

	eps := endpoints.AllEndpoints()
	wg := sync.WaitGroup{}
	for _, ep := range eps {
		wg.Add(1)
		go func(ep endpoints.Endpoint) {
			defer wg.Done()
			tableCtx, _ := tag.New(deployCtx,
				tag.Insert(metrics.KeyVersion, string(ep.Version)),
				tag.Insert(metrics.KeyResource, ep.Resource))
			url, err := ep.URL()
			if err != nil {
				log.Printf("error getting %v stats: %v", &ep, err)
				return
			}

			log.Printf("Fetching %s..", url)
			urlStats, err := freshstats.GetURLStats(url)
			if err != nil {
				log.Printf("error getting %v stats %q: %v", &ep, url, err)
				return
				// XXX FUTURE report a count of errors
			}

			stats.Record(tableCtx, metrics.FileLengthBytes.M(int64(urlStats.FileLengthBytes)))
			stats.Record(tableCtx, metrics.FileLengthJSONItems.M(int64(urlStats.FileLengthJSONItems)))

			if !urlStats.LastModified.IsZero() {
				stats.Record(tableCtx, metrics.LastModified.M(urlStats.LastModified.Unix()))
				ago := time.Since(urlStats.LastModified).Seconds()
				stats.Record(tableCtx, metrics.LastModifiedAge.M(ago))
			}
		}(ep)
	}
	wg.Wait()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "done")
}
