package legacy

import "github.com/CAVaccineInventory/airtable-export/pkg/filter"

// NOTE: this is deprecated.
var countiesAllowList = map[string]struct{}{
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

func Counties(jsonMap []map[string]interface{}) []map[string]interface{} {
	return filter.FilterToAllowedKeys(jsonMap, countiesAllowList)
}
