package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/CAVaccineInventory/airtable-export/monitoring/freshcf/pkg/stats"
)

func ExportJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, cxl := context.WithTimeout(ctx, requestTimeout)
	defer cxl()

	resultChan := stats.AllResponses(ctx)

	results := make(map[string]stats.ExportedJSONFileStats)
	for len(resultChan) != 0 {
		result := <-resultChan
		if result.Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error getting stats for %s: %v", result.URL, result.Err)
			return
		}
		results[result.URL] = result.Stats
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
