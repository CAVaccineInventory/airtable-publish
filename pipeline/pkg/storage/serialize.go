package storage

import (
	"bytes"
	"encoding/json"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/metadata"
)

func Serialize(jd metadata.JSONData) (*bytes.Buffer, error) {
	unsanitizedJSON, err := json.Marshal(jd)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	json.HTMLEscape(buf, unsanitizedJSON)
	return buf, nil
}
