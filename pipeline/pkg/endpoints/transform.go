package endpoints

import (
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
)

func RemoveDisallowedFields(jsonMap airtable.TableContent, allowed map[string]int) {
	for i := range jsonMap {
		for k := range jsonMap[i] {
			if _, found := allowed[k]; !found {
				delete(jsonMap[i], k)
			}
		}
	}
}
