package counties

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/honeycombio/beeline-go"
)

var (
	// NOTE: This map is preliminary and has not yet been vetted by the
	// deciders. It's here for now as an example of what this kind of map might
	// look like.
	v2Map = map[string]string{
		"County":                              "name",
		"County vaccination reservations URL": "reservationsURL",
		"Facebook Page":                       "facebookURL",
		"Notes":                               "notes",
		"Official volunteering opportunities": "officialVolunteering",
		"Total reports":                       "totalReports",
		"Twitter Page":                        "twitterURL",
		"Vaccine info URL":                    "vaccineInfoURL",
		"Vaccine locations URL":               "vaccineLocationsURL",
		"Yeses":                               "yesses",
	}
)

func V2(ctx context.Context, tables *airtable.Tables) (types.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "endpoints.counties.V2")
	defer span.Send()

	rawTable, err := tables.GetCounties(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Counties table: %w", err)
	}
	filteredTable, err := filter.RemapToAllowedKeys(rawTable, v2Map)
	if err != nil {
		return nil, fmt.Errorf("RemapToAllowedKeys: %w", err)
	}
	return filteredTable, nil
}
