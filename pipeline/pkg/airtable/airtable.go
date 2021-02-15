package airtable

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/config"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/secrets"
	beeline "github.com/honeycombio/beeline-go"
)

type TableContent []map[string]interface{}

// Unmarshals and returns JSON stored at the given filePath.
func ObjectFromFile(ctx context.Context, tableName string, filePath string) (TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "airtable.ObjectFromFile")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		err = fmt.Errorf("couldn't read file %s: %w", filePath, err)
		beeline.AddField(ctx, "error", err)
		return nil, err
	}
	log.Printf("[%s] Read %d bytes from disk (%s).\n", tableName, len(b), filePath)

	jsonMap := make([]map[string](interface{}), 0)
	err = json.Unmarshal([]byte(b), &jsonMap)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return nil, err
	}
	return jsonMap, err
}

// Represents a single row as returned from the Airtable API
type responseRow struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// Represents the full response as returned by the Airtable API
type responseData struct {
	Offset  string        `json:"offset"`
	Records []responseRow `json:"records"`
}

// Makes a single request to the Airtable endpoint; returns the new
// rows, next offset, and error.  Wraps fetchRowsActual with tracing.
func fetchRows(ctx context.Context, tableName string, offset string) (TableContent, string, error) {
	ctx, span := beeline.StartSpan(ctx, "airtable.fetchRows")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)
	beeline.AddField(ctx, "offset", offset)

	rows, offset, err := fetchRowsActual(ctx, tableName, offset)
	if err != nil {
		err = fmt.Errorf("failed to fetch table %s: %w", tableName, err)
		beeline.AddField(ctx, "error", err)
	}
	return rows, offset, err
}

// Makes a single request to the Airtable endpoint; returns the new
// rows, next offset, and error.
func fetchRowsActual(ctx context.Context, tableName string, offset string) (TableContent, string, error) {
	url := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", config.AirtableID, tableName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return TableContent{}, offset, err
	}

	airtableSecret, err := secrets.Get(ctx, secrets.AirtableSecret)
	if err != nil {
		return TableContent{}, offset, fmt.Errorf("Failed to fetch airtable secret: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", airtableSecret))

	q := req.URL.Query()
	q.Add("offset", offset)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TableContent{}, offset, err
	}
	beeline.AddField(ctx, "statusCode", resp.StatusCode)
	if resp.StatusCode == http.StatusTooManyRequests {
		time.Sleep(200 * time.Millisecond)
		return TableContent{}, offset, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TableContent{}, offset, fmt.Errorf("Got response code %d", resp.StatusCode)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TableContent{}, offset, err
	}

	d := json.NewDecoder(strings.NewReader(string(bytes)))
	rd := responseData{}
	if err = d.Decode(&rd); err != nil {
		return TableContent{}, offset, err
	}
	// Max page size is 100
	rows := make(TableContent, 0, 100)
	for _, row := range rd.Records {
		row.Fields["id"] = row.ID // synthetic "id" field based on Airtable ID takes precedence over any field that might be named "id".
		rows = append(rows, row.Fields)

	}
	return rows, rd.Offset, nil

}

// Downloads a table from Airtable, and returns the unmarshaled data
// from it.  Airtable limits to paging 100 rows per request, 5
// requests per second, so this may take a large number of requests.
func Download(ctx context.Context, tableName string) (TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "airtable.Download")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	jsonMap := make(TableContent, 0)
	offset := ""
	for {
		rows, nextOffset, err := fetchRows(ctx, tableName, offset)
		if err != nil {
			return TableContent{}, err
		}
		jsonMap = append(jsonMap, rows...)
		if nextOffset == "" {
			break
		}
		offset = nextOffset
		time.Sleep(200 * time.Millisecond)
	}
	return jsonMap, nil
}
