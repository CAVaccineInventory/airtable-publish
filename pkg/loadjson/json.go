package loadjson

import (
	"encoding/json"
	"github.com/CAVaccineInventory/airtable-export/pkg/table"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
)

func TableFromJson(filePath string) (table.Table, error) {
	b, readErr := ioutil.ReadFile(filePath)
	if readErr != nil {
		return nil, errors.Wrapf(readErr, "couldn't read file %s", filePath)
	}
	log.Printf("Read %d bytes from disk (%s).\n", len(b), filePath)

	jsonMap := make([]map[string](interface{}), 0)
	marshalErr := json.Unmarshal([]byte(b), &jsonMap)
	return jsonMap, marshalErr
}
