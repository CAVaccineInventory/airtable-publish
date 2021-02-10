package filter

import (
	"errors"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
)

// SameKeyMapping takes a list of keys, and returns a map of key -> key.
// It is intended as a helper for RemapToAllowedKeys, when the key names are the same.
func SameKeyMapping(keys []string) map[string]string {
	mapping := make(map[string]string, len(keys))
	for _, k := range keys {
		mapping[k] = k
	}
	return mapping
}

// RemapToAllowedKeys takes a slice of KV objects, and a map of allowed key names => output names.
// For each object in the list, it removes each KV pair where the key is not in allowedKeys,
// then returns this result.
func RemapToAllowedKeys(raw airtable.TableContent, fields map[string]string) (airtable.TableContent, error) {
	filtered := make([]map[string]interface{}, len(raw))

	for i := range raw {
		filtered[i] = map[string]interface{}{}
		for k, v := range raw[i] {
			if _, ok := fields[k]; ok {
				filtered[i][fields[k]] = v
			}
		}
	}

	return filtered, checkFields(filtered, fields)
}

// ErrMissingField represents the case when a field is missing from the output.
var ErrMissingField = errors.New("missing field")

// checkFields makes sure that every specified field shows up at least once.  It
// does not check that every record has every field.
func checkFields(data airtable.TableContent, fields map[string]string) error {
	var seen = make(map[string]struct{}, len(fields))

	for i := range data {
		for k := range data[i] {
			seen[k] = struct{}{}
		}
	}

	for _, f := range fields {
		if _, ok := seen[f]; !ok {
			return fmt.Errorf("%w: %q missing", ErrMissingField, f)
		}
	}

	return nil
}
