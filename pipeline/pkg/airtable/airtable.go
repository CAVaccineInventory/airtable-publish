package airtable

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/config"
	beeline "github.com/honeycombio/beeline-go"
	"go.opencensus.io/stats"
)

type TableContent []map[string]interface{}

func ObjectFromFile(ctx context.Context, tableName string, filePath string) (TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "airtable.ObjectFromFile")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %s: %w", filePath, err)
	}
	log.Printf("[%s] Read %d bytes from disk (%s).\n", tableName, len(b), filePath)

	jsonMap := make([]map[string](interface{}), 0)
	err = json.Unmarshal([]byte(b), &jsonMap)
	return jsonMap, err
}

// Download downloads a table from Airtable, and returns the
// unmarshaled data from it.
func Download(ctx context.Context, tableName string) (TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "airtable.Download")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	airtableSecret := os.Getenv(config.AirtableSecretEnvKey)

	// TODO: consider doing this in Go directly.
	tempDir, err := ioutil.TempDir("", tableName)
	defer os.RemoveAll(tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to make temp directory: %w", err)
	}

	log.Printf("[%s] Shelling out to exporter...\n", tableName)
	start := time.Now()
	cmd := exec.CommandContext(ctx, "/usr/bin/airtable-export", "--json", tempDir, config.AirtableID, tableName, "--key", airtableSecret)
	output, err := cmd.CombinedOutput()
	stats.Record(ctx, FetchLatency.M(time.Since(start).Seconds()))
	if err != nil {
		log.Println("Output from failed airtable-export:\n" + string(output))
		return nil, fmt.Errorf("failed to run airtable-export: %w", err)
	}
	outputFile := path.Join(tempDir, tableName+".json")

	jsonMap, err := ObjectFromFile(ctx, tableName, outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json in %s: %w", outputFile, err)
	}
	return jsonMap, nil
}
