package main

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
	"path"
)

// fetchAirtableTable dumps a a table from Airtable as JSON on disk.
// Returns the path of the output file, and an error.
func fetchAirtableTable(tableName string) (string, error) {
	airtableSecret := os.Getenv(airtableSecretEnvKey)

	// TODO: consider doing this in Go directly.
	log.Println("Shelling out to exporter...")
	cmd := exec.Command("/usr/bin/airtable-export", "--json", tempDir, airtableId, tableName, "--key", airtableSecret)
	output, exportErr := cmd.CombinedOutput()
	if exportErr != nil {
		log.Println(string(output))
		return "", errors.Wrap(exportErr, "failed to run airtable-export")
	}
	return path.Join(tempDir, tableName+".json"), nil
}
