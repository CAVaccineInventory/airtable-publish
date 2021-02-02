package stats

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints"
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
func GetURLStats(url string) (ExportedJSONFileStats, error) {
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

type Response struct {
	URL   string
	Stats ExportedJSONFileStats
	Err   error
}

func AllResponses() chan Response {
	eps := endpoints.AllEndpoints()
	resultChan := make(chan Response, len(eps))
	wg := sync.WaitGroup{}
	for _, ep := range eps {
		wg.Add(1)
		go func(ep endpoints.Endpoint) {
			defer wg.Done()
			url, err := ep.URL()
			if err != nil {
				log.Printf("error getting %v stats: %v", &ep, err)
				resultChan <- Response{URL: url, Stats: ExportedJSONFileStats{}, Err: err}
				return
			}

			log.Printf("Fetching %s..", url)
			stats, err := GetURLStats(url)
			resultChan <- Response{URL: url, Stats: stats, Err: err}
		}(ep)
	}
	wg.Wait()
	return resultChan
}
