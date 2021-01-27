package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

var allowKeys = map[string]map[string]int{
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

func Sanitize(jsonMap []map[string]interface{}, tableName string) (*bytes.Buffer, error) {
	keys, ok := allowKeys[tableName]

	if !ok {
		return nil, fmt.Errorf("ERROR: unsupported kind of export: %s", tableName)
	}

	for i := range jsonMap {
		for k := range jsonMap[i] {
			if _, found := keys[k]; !found {
				delete(jsonMap[i], k)
			}
		}
	}
	log.Printf("[%s] Cleaned %d elements.\n", tableName, len(jsonMap))

	unsanitizedJSON, err := json.Marshal(jsonMap)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	json.HTMLEscape(buf, unsanitizedJSON)

	return buf, nil
}
