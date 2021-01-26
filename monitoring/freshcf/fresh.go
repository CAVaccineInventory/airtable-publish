// Package fresh contains a Cloud Function for checking URL freshness.
package freshcf

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/locations"
)

type ExportedJSONFileStats struct {
	lastModifiedAgeSeconds int
	fileLengthJsonItems    int
	fileLengthBytes        int
}

// ExportedJSONFileStats is always filled out to the best of our
// ability. error will be true if there was a fatal error along the
// way, in which case values that are not meaningful given the error
// will be 0.
func getURLStats(url string) (ExportedJSONFileStats, error) {
	output := ExportedJSONFileStats{}

	resp, err := http.Get(url)
	if err != nil {
		return output, errors.Wrapf(err, "fetching: %q", url)
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
		output.lastModifiedAgeSeconds = int(ago)
		log.Printf("Last-Modified: %v (%0.fs ago)", when, ago)
	}

	// get the contents.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return output, errors.Wrapf(err, "reading: %q", url)
	}
	output.fileLengthBytes = len(body)

	// parse the body
	var jsonBody []interface{}
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		log.Printf("failed to parse JSON %v", err)
	} else {
		output.fileLengthJsonItems = len(jsonBody)
	}

	return output, nil
}

// CheckFreshness checks the freshness of the Locations.json
func CheckFreshness(w http.ResponseWriter, r *http.Request) {
	threshold_age := 600
	threshold_items := 10
	threshold_length := 1000

	var thr = ""
	thr = os.Getenv("THRESHOLD_AGE")
	if thr != "" {
		if t, err := strconv.Atoi(thr); err == nil {
			threshold_age = t
		}
	}

	thr = os.Getenv("THRESHOLD_ITEMS")
	if thr != "" {
		if t, err := strconv.Atoi(thr); err == nil {
			threshold_items = t
		}
	}

	thr = os.Getenv("THRESHOLD_LENGTH")
	if thr != "" {
		if t, err := strconv.Atoi(thr); err == nil {
			threshold_length = t
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

	if stats.lastModifiedAgeSeconds > threshold_age {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "last modified is too old: %d < %d", stats.lastModifiedAgeSeconds, threshold_age)
		return
	}

	if stats.fileLengthBytes < threshold_length {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "file body too short: %d < %d", stats.fileLengthBytes, threshold_length)
		return
	}

	if stats.fileLengthJsonItems < threshold_items {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "json list too short: %d < %d", stats.fileLengthJsonItems, threshold_items)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
