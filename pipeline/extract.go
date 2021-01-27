package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
)

// fetchAirtableTable dumps a a table from Airtable as JSON on disk.
func fetchAirtableTable(ctx context.Context, tempDir string, tableName string) (string, error) {
	airtableSecret := os.Getenv(airtableSecretEnvKey)

	// TODO: consider doing this in Go directly.
	log.Printf("[%s] Shelling out to exporter...\n", tableName)
	cmd := exec.CommandContext(ctx, "/usr/bin/airtable-export", "--json", tempDir, airtableID, tableName, "--key", airtableSecret)
	output, exportErr := cmd.CombinedOutput()
	if exportErr != nil {
		log.Println(string(output))
		return "", errors.Wrap(exportErr, "failed to run airtable-export")
	}
	j := path.Join(tempDir, tableName+".json")
	return j, nil
}
