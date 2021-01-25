// Package fresh contains a Cloud Function for checking URL freshness.
package freshcf

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var urls = map[string]string{
	"prod": "https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json",
	"staging": "https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync-staging/Locations.json",
}

// CheckFreshness checks the freshness of the Locations.json
func CheckFreshness(w http.ResponseWriter, r *http.Request) {
	thresh := 600

	thr := os.Getenv("THRESHOLD")
	if thr != "" {
		if t, err := strconv.Atoi(thr); err == nil {
			thresh = t
		}
	}

	deploy := os.Getenv("DEPLOY")
	url, found := urls[deploy];
	if !found {
		url = urls["prod"]
	}

	resp, err := http.Head(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error retrieving URL %q: %v", url, err)
		return
	}

	lu := resp.Header.Get("Last-Modified")
	if lu == "" {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "no Last-Modified header for URL %q: %v", url, err)
		return
	}

	when, err := time.Parse(time.RFC1123, lu)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error parsing %q as a time: %v", lu, err)
		return
	}
	ago := time.Now().Sub(when).Seconds()
	log.Printf("Last-Modified: %v (%0.fs ago)", when, ago)

	if int(ago) > thresh {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%q  (%.0f seconds ago) is too old", lu, ago)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
