// Package fresh contains a Cloud Function for checking URL freshness.
package freshcf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/locations"
)

type ExportedJSONFileStats struct {
	LastModifiedAgeSeconds int `json:"last_modified_age_seconds"`
	FileLengthJSONItems    int `json:"file_length_json_items"`
	FileLengthBytes        int `json:"file_length_bytes"`
}

const NOVALUE int = -1

// ExportedJSONFileStats is always filled out to the best of our
// ability. error will be true if there was a fatal error along the
// way, in which case values that are not meaningful given the error
// will be set to NOVALUE.
func getURLStats(url string) (ExportedJSONFileStats, error) {
	output := ExportedJSONFileStats{}
	output.LastModifiedAgeSeconds = NOVALUE
	output.FileLengthJSONItems = NOVALUE
	output.FileLengthBytes = NOVALUE

	resp, err := http.Get(url)
	if err != nil {
		return output, fmt.Errorf("fetching: %q: %w", url, err)
	}

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
		ago := time.Since(when).Seconds()
		output.LastModifiedAgeSeconds = int(ago)
		log.Printf("Last-Modified: %v (%0.fs ago)", when, ago)
	}

	// get the contents.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return output, fmt.Errorf("reading: %q: %w", url, err)
	}
	output.FileLengthBytes = len(body)

	// parse the body
	var jsonBody []interface{}
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		log.Printf("failed to parse JSON %v", err)
	} else {
		output.FileLengthJSONItems = len(jsonBody)
	}

	return output, nil
}

func ExportJSON(w http.ResponseWriter, r *http.Request) {
	deploy := os.Getenv("DEPLOY")
	url, found := urls[deploy]
	if !found {
		url = urls["prod"]
	}

	stats, err := getURLStats(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error getting stats %q: %v", url, err)
		return
	}

	jsonBytes, err := json.Marshal(stats)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error marshalling json %q: %v", url, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(jsonBytes))
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

	url := locations.GetExportBaseURL() + "/Locations.json"
	resp, err := http.Head(url)

	stats, err := getURLStats(url)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error getting stats %q: %v", url, err)
		return
	}

	if stats.LastModifiedAgeSeconds > thresholdAge || stats.LastModifiedAgeSeconds == NOVALUE {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "last modified is too old: %d < %d", stats.LastModifiedAgeSeconds, thresholdAge)
		return
	}

	if stats.FileLengthBytes < thresholdLength || stats.FileLengthBytes == NOVALUE {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "file body too short: %d < %d", stats.FileLengthBytes, thresholdLength)
		return
	}

	if stats.FileLengthJSONItems < thresholdItems || stats.FileLengthJSONItems == NOVALUE {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "json list too short: %d < %d", stats.FileLengthJSONItems, thresholdItems)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
