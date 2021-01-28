package sanitize

import (
	"bytes"
	"encoding/json"
)

func Sanitize(raw interface{}) (*bytes.Buffer, error) {
	unsanitized, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	json.HTMLEscape(buf, unsanitized)

	return buf, nil
}
