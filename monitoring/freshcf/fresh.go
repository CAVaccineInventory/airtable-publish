// Package fresh contains a Cloud Function for checking URL freshness.
package freshcf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/metrics"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

type ExportedJSONFileStats struct {
	LastModified        time.Time `json:"last_modified"`
	FileLengthJSONItems int       `json:"file_length_json_items"`
	FileLengthBytes     int       `json:"file_length_bytes"`
}

// ExportedJSONFileStats is always filled out to the best of our
// ability. error will be true if there was a fatal error along the
// way, in which case values that are not meaningful given the error
// will be 0.
func getURLStats(url string) (ExportedJSONFileStats, error) {
	output := ExportedJSONFileStats{}

	resp, err := http.Get(url)
	if err != nil {
		return output, fmt.Errorf("fetching: %q: %w", url, err)
	}
	// Always clean up the response body.
	defer resp.Body.Close()

	// First, check if we got a successful HTTP response. If not,
	// Last-Modified is not valid/useful.
	if resp.StatusCode != 200 {
		return output, errors.New("non-200 status code")
	}

	// Next, header checks.
	lu := resp.Header.Get("Last-Modified")
	when, err := time.Parse(time.RFC1123, lu)
	if err != nil {
		log.Printf("invalid last-modified %v", err)
	} else {
		output.LastModified = when
	}

	// get the contents.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return output, fmt.Errorf("reading: %q: %w", url, err)
	}
	output.FileLengthBytes = len(body)

	// parse the body
	var jsonData interface{}
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		return output, err
	}
	if listData, ok := jsonData.([]interface{}); ok {
		output.FileLengthJSONItems = len(listData)
		return output, nil
	}

	if mapData, ok := jsonData.(map[string]interface{}); ok {
		if listPart, ok := mapData["content"].([]interface{}); ok {
			output.FileLengthJSONItems = len(listPart)
			return output, nil
		}
	}

	return output, errors.New("Unknown JSON structure")
}

type StatsResponse struct {
	url   string
	stats ExportedJSONFileStats
	err   error
}

func AllResponses() chan StatsResponse {
	eps := endpoints.AllEndpoints()
	resultChan := make(chan StatsResponse, len(eps))
	wg := sync.WaitGroup{}
	for _, ep := range eps {
		wg.Add(1)
		go func(ep endpoints.Endpoint) {
			defer wg.Done()
			url, err := ep.URL()
			if err != nil {
				log.Printf("error getting %v stats: %v", &ep, err)
				resultChan <- StatsResponse{url: url, stats: ExportedJSONFileStats{}, err: err}
				return
			}

			log.Printf("Fetching %s..", url)
			stats, err := getURLStats(url)
			resultChan <- StatsResponse{url: url, stats: stats, err: err}
		}(ep)
	}
	wg.Wait()
	return resultChan
}

func ExportJSON(w http.ResponseWriter, r *http.Request) {
	resultChan := AllResponses()

	results := make(map[string]ExportedJSONFileStats)
	for len(resultChan) != 0 {
		result := <-resultChan
		if result.err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error getting stats for %s: %v", result.url, result.err)
			return
		}
		results[result.url] = result.stats
	}

	jsonBytes, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error marshalling json: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Printf("Error writing to client %v", err)
	}
}

// CheckFreshness checks the freshness of the Locations.json
func CheckFreshness(w http.ResponseWriter, r *http.Request) {
	thresholdAge := 600
	thresholdItems := 10
	thresholdLength := 1000

	var thr string
	thr = os.Getenv("THRESHOLD_AGE")
	if thr != "" {
		if t, err := strconv.Atoi(thr); err == nil {
			thresholdAge = t
		}
	}

	thr = os.Getenv("THRESHOLD_ITEMS")
	if thr != "" {
		if t, err := strconv.Atoi(thr); err == nil {
			thresholdItems = t
		}
	}

	thr = os.Getenv("THRESHOLD_LENGTH")
	if thr != "" {
		if t, err := strconv.Atoi(thr); err == nil {
			thresholdLength = t
		}
	}

	resultChan := AllResponses()
	errs := make([]string, 0)
	for len(resultChan) != 0 {
		result := <-resultChan
		if result.err != nil {
			errs = append(errs, fmt.Sprintf("%s\nerror fetching: %s", result.url, result.err))
			continue
		}

		stats := result.stats
		if stats.LastModified.IsZero() {
			errs = append(errs, fmt.Sprintf("%s\ninvalid last modified header", result.url))
			continue
		}

		ago := int(time.Since(stats.LastModified).Seconds())
		if ago > thresholdAge {
			errs = append(errs, fmt.Sprintf("%s\nlast modified is too old: %d < %d", result.url, ago, thresholdAge))
			continue
		}

		if stats.FileLengthBytes < thresholdLength {
			errs = append(errs, fmt.Sprintf("%s\nfile body too short: %d < %d", result.url, stats.FileLengthBytes, thresholdLength))
			continue
		}

		if stats.FileLengthJSONItems < thresholdItems {
			errs = append(errs, fmt.Sprintf("%s\njson list too short: %d < %d", result.url, stats.FileLengthJSONItems, thresholdItems))
			continue
		}
	}

	if len(errs) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Errors:\n\n%v", strings.Join(errs, "\n\n"))
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	}
}

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
			urlStats, err := getURLStats(url)
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
