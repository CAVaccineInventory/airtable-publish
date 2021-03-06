package legacy

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
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

func Locations(ctx context.Context, tables *airtable.Tables) (types.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "endpoints.legacy.Locations")
	defer span.Send()

	rawTable, err := tables.GetLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Locations table: %w", err)
	}
	fields := []string{
		"Address",
		"Affiliation",
		"Appointment scheduling instructions",
		"Availability Info",
		"County",
		"Has Report",
		"Latest report",
		"Latest report notes",
		"Latest report yes?",
		"Latitude",
		"Location Type",
		"Longitude",
		"Name",
		"vaccinefinder_location_id",
		"vaccinespotter_location_id",
		"google_places_id",
	}

	filteredTable, err := filter.Transform(rawTable, filter.WithFieldSlice(fields))
	if err != nil {
		return nil, fmt.Errorf("Transform: %w", err)
	}

	return filteredTable, nil
}

func Counties(ctx context.Context, tables *airtable.Tables) (types.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "endpoints.legacy.Counties")
	defer span.Send()

	rawTable, err := tables.GetCounties(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Counties table: %w", err)
	}
	fields := []string{
		"County",
		"County vaccination reservations URL",
		"Facebook Page",
		"Notes",
		"Official volunteering opportunities",
		"Total reports",
		"Twitter Page",
		"Vaccine info URL",
		"Vaccine locations URL",
		"Yeses",
		"age_floor_without_restrictions",
	}

	filteredTable, err := filter.Transform(rawTable, filter.WithFieldSlice(fields))
	if err != nil {
		return nil, fmt.Errorf("Transform: %w", err)
	}

	return filteredTable, nil
}
