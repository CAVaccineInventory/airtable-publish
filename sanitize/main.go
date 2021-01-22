package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

// Extracted from data.js using "get_required_fields_for_site.py".
var allowKeys = map[string]int{
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
}

func main() {

	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Print(err)
	}

	jsonMap := make([]map[string](interface{}), 0)
	err = json.Unmarshal([]byte(b), &jsonMap)
	if err != nil {
		log.Printf("ERROR: fail to unmarshal json, %s", err.Error())
	}

	repack := make([]map[string](interface{}), 0)
	for _, element := range jsonMap {
		var sanitized_element = make(map[string](interface{}))
		for k, v := range element {
			_, ok := allowKeys[k]
			if ok {
				sanitized_element[k] = v
			}
		}
		repack = append(repack, sanitized_element)
	}

	unsanitizedJson, err := json.Marshal(repack)
	buf := &bytes.Buffer{}
	json.HTMLEscape(buf, unsanitizedJson)
	fmt.Println(buf)
}
