package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	beeline "github.com/honeycombio/beeline-go"
)

// fetchAirtableTable dumps a a table from Airtable as JSON on disk.
func fetchAirtableTable(ctx context.Context, tempDir string, tableName string) (string, error) {
	ctx, span := beeline.StartSpan(ctx, "fetch-airtable-table")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	airtableSecret := os.Getenv(airtableSecretEnvKey)

	// TODO: consider doing this in Go directly.
	log.Printf("[%s] Shelling out to exporter...\n", tableName)
	cmd := exec.CommandContext(ctx, "/usr/bin/airtable-export", "--json", tempDir, airtableID, tableName, "--key", airtableSecret)
	output, exportErr := cmd.CombinedOutput()
	if exportErr != nil {
		log.Println("Output from failed airtable-export:\n" + string(output))
		return "", fmt.Errorf("failed to run airtable-export: %w", exportErr)
	}
	j := path.Join(tempDir, tableName+".json")
	return j, nil
}
