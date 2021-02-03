package storage

import (
	"bytes"
	"encoding/json"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/metadata"
)

// JSON marshals the data, then HTML-escapes it so it could possibly
// be inlined within <script> tags (though we do not currently do so).
func Serialize(jd metadata.JSONData) (*bytes.Buffer, error) {
	unsanitizedJSON, err := json.Marshal(jd)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	json.HTMLEscape(buf, unsanitizedJSON)
	return buf, nil
}
