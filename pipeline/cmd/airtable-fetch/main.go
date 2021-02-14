// airtable-fetch is a utility program to download the raw airtable data used as
// input for the pipeline.  Useful for exploring the raw data to see what the
// input is.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/secrets"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

var (
	table = flag.String("table", "Locations", "table to fetch (Locations, Counties, Providers, etc.)")
)

func main() {
	flag.Parse()

	if *table == "" {
		log.Fatal("required flag --table not specified")
	}

	ctx := context.Background()
	secrets.RequireAirtableSecret()
	tables := airtable.NewTables()

	var data types.TableContent
	var err error

	switch strings.ToLower(*table) {
	case "locations":
		data, err = tables.GetLocations(ctx)
	case "counties":
		data, err = tables.GetCounties(ctx)
	case "proivders":
		data, err = tables.GetProviders(ctx)
	default:
		log.Fatalf("invalid value for --table: %v", *table)
	}
	if err != nil {
		log.Fatalf("error fetching %v: %v", *table, err)
	}

	bs, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		log.Fatalf("error marshaling JSON: %v", err)
	}

	os.Stdout.Write(bs)

}
