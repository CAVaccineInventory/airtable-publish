package locations

import (
	"github.com/CAVaccineInventory/airtable-export/pkg/loadjson"
	"github.com/CAVaccineInventory/airtable-export/pkg/table"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocationsV1(t *testing.T) {
	tableJson, err := loadjson.TableFromJson("../../../test_data/locations_reduced.json")
	assert.NoError(t, err)

	response, err := LocationsV1.GenerateResponse(map[string]table.Table{"Locations": tableJson})
	assert.NoError(t, err)
	listResponse := response.Content.([]map[string]interface{})

	assert.NotEqual(t, 0, len(listResponse))
	for _, location := range listResponse {
		_, found := location["Name"]
		assert.True(t, found, "should find name")

		_, found = location["Phone number"]
		assert.False(t, found, "shouldn't find phone number")
	}
}
