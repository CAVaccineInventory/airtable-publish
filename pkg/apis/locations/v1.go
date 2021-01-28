package locations

import (
	"errors"
	"github.com/CAVaccineInventory/airtable-export/pkg/apis/apimeta"
	"github.com/CAVaccineInventory/airtable-export/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pkg/table"
)

var locationsAllowList = map[string]struct{}{
	"Address":                             {},
	"Appointment scheduling instructions": {},
	"Availability Info":                   {},
	"County":                              {},
	"Has Report":                          {},
	"Latest report":                       {},
	"Latest report notes":                 {},
	"Latest report yes?":                  {},
	"Latitude":                            {},
	"Location Type":                       {},
	"Longitude":                           {},
	"Name":                                {},
}

// LocationsV1 defines the endpoint /locations/v1/locations.
var LocationsV1 = apimeta.EndpointDefinition{
	Group: group,
	Kind:  "locations",
	GenerateResponse: func(tables map[string]table.Table) (apimeta.List, error) {
		locationsTable, found := tables["Locations"]
		if !found {
			return apimeta.List{}, errors.New("table not found")
		}

		filteredLocations := filter.FilterToAllowedKeys(locationsTable, locationsAllowList)

		return apimeta.List{
			Metadata: apimeta.Metadata{
				ApiVersion: apimeta.ApiVersion{
					Major:     1,
					Minor:     0,
					Stability: "unstable",
				},
				Contact:     apimeta.DefaultContact,
				UsageNotice: "TODO",
			},
			Content: filteredLocations,
		}, nil
	},
}
