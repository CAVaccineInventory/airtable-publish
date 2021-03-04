package locations

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/honeycombio/beeline-go"
)

var (
	v2Map = map[string]string{
		"Address":                             "address",
		"Affiliation":                         "affiliation",
		"Appointment scheduling instructions": "appointment_scheduling_instructions",
		"Availability Info":                   "availability_info",
		"County":                              "county",
		"Has Report":                          "has_report",
		"Latest report":                       "latest_report",
		"Latest report notes":                 "latest_report_notes",
		"Latest report yes?":                  "latest_report_is_yes",
		"Latitude":                            "latitude",
		"Location Type":                       "location_type",
		"Longitude":                           "longitude",
		"Name":                                "name",
		"vaccinefinder_location_id":           "vaccinefinder_location_id",
		"google_places_id":                    "google_places_id",
	}
)

// locationsTransformer is a filter.Munger,
// which transforms the contents of a given item/row in
func locationsTransformer(in map[string]interface{}) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	for k, v := range in {
		// Convert appointment_scheduling_instructions from []string{"value"} to "value".
		if k == "appointment_scheduling_instructions" {
			cast, ok := in[k].([]string)
			// Drop the field if it deviates from expectations.
			// TODO: this seems like an opportune place to count warnings or something.
			if ok && len(cast) == 1 {
				out[k] = cast[0]
			}
		} else { // Keep all other keys as-is.
			out[k] = v
		}
	}

	return out, nil
}

func V2(ctx context.Context, tables *airtable.Tables) (types.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "endpoints.locations.V2")
	defer span.Send()

	rawTable, err := tables.GetLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Locations table: %w", err)
	}
	filteredTable, err := filter.Transform(rawTable, filter.WithFieldMap(v2Map), filter.WithMunger(locationsTransformer))
	if err != nil {
		return nil, fmt.Errorf("Transform: %w", err)
	}

	return filteredTable, nil
}
