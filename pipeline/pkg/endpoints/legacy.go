package endpoints

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	"github.com/honeycombio/beeline-go"
)

/**
https://discord.com/channels/799147121357881364/799508020098629632/801607940518707270

So the SF Chronicle wants to consume our "api", and these are the fields they're requesting:

Name
Location
County
Availability info
Scheduling info
Latest report time


https://discord.com/channels/799147121357881364/799508020098629632/801611160520097823
Okay, for those doing exporter stuff, for the following fields: "Name", "Address", "Latitude", "Longitude", "County", "Availability Info", "Latest report yes?", "How to schedule appointments", "Latest Report"

 - please don't break them for now (they're used by the map, but the sf chronicle might also use them, so @ me if you plan to break them
 - perhaps we should give some of these better names and stick that in Locations-v2.json
*/

var legacyAllowKeys = map[string]map[string]int{
	// Extracted from data.js using "get_required_fields_for_site.py".
	"Locations": {
		"Address":                             1,
		"Appointment scheduling instructions": 1,
		"Availability Info":                   1,
		"County":                              1,
		"Has Report":                          1,
		"Latest report":                       1,
		"Latest report notes":                 1,
		"Latest report yes?":                  1,
		"Latitude":                            1,
		"Location Type":                       1,
		"Longitude":                           1,
		"Name":                                1,
	},
	"Counties": {
		"County":                              1,
		"Vaccine info URL":                    1,
		"Vaccine locations URL":               1,
		"Notes":                               1,
		"Total reports":                       1,
		"Yeses":                               1,
		"Official volunteering opportunities": 1,
		"Facebook Page":                       1,
	},
}

func GenerateV1Locations(ctx context.Context, getTable generator.TableFetchFunc) (airtable.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "generator.GenerateV1Locations")
	defer span.Send()

	jsonMap, err := getTable(ctx, "Locations")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Locations table: %w", err)
	}
	RemoveDisallowedFields(jsonMap, legacyAllowKeys["Locations"])
	return jsonMap, nil
}

func GenerateV1Counties(ctx context.Context, getTable generator.TableFetchFunc) (airtable.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "generator.GenerateV1Counties")
	defer span.Send()

	jsonMap, err := getTable(ctx, "Counties")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Locations table: %w", err)
	}
	RemoveDisallowedFields(jsonMap, legacyAllowKeys["Counties"])
	return jsonMap, nil
}
