package legacy

import "github.com/CAVaccineInventory/airtable-export/pkg/filter"

// NOTE: this is deprecated.
// At the time of merging (and likely not for much longer)
// it is used by the website.
// The SF Chronicle is also using it in an exploratory fashion.
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

func Locations(jsonMap []map[string]interface{}) []map[string]interface{} {
	return filter.FilterToAllowedKeys(jsonMap, locationsAllowList)
}
