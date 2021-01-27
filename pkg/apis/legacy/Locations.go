package legacy

import "github.com/CAVaccineInventory/airtable-export/pkg/filter"

var locationsALlowList = map[string]struct{}{
	"Address": {},
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

func Locations(jsonMap []map[string]interface{}) []map[string]interface{} {
	filter.FilterToAllowedKeys(jsonMap, locationsALlowList)
	return jsonMap
}
