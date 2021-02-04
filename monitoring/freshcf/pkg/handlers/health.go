package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/stats"
)

func Health(w http.ResponseWriter, r *http.Request) {
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

	resultChan := stats.AllResponses(r.Context())
	errs := make([]string, 0)
	for len(resultChan) != 0 {
		result := <-resultChan
		if result.Err != nil {
			errs = append(errs, fmt.Sprintf("%s\nerror fetching: %s", result.URL, result.Err))
			continue
		}

		stats := result.Stats
		if stats.LastModified.IsZero() {
			errs = append(errs, fmt.Sprintf("%s\ninvalid last modified header", result.URL))
			continue
		}

		ago := int(time.Since(stats.LastModified).Seconds())
		if ago > thresholdAge {
			errs = append(errs, fmt.Sprintf("%s\nlast modified is too old: %d < %d", result.URL, ago, thresholdAge))
			continue
		}

		if stats.FileLengthBytes < thresholdLength {
			errs = append(errs, fmt.Sprintf("%s\nfile body too short: %d < %d", result.URL, stats.FileLengthBytes, thresholdLength))
			continue
		}

		if stats.FileLengthJSONItems < thresholdItems {
			errs = append(errs, fmt.Sprintf("%s\njson list too short: %d < %d", result.URL, stats.FileLengthJSONItems, thresholdItems))
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
