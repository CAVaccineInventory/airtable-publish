package filter

import (
	"errors"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

// ErrMissingField represents the case when a field is missing from the output.
var ErrMissingField = errors.New("missing field")

// checkFields makes sure that every specified field shows up at least once.  It
// does not check that every record has every field.
func checkFields(data types.TableContent, fields map[string]string) error {
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
