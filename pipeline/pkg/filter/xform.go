package filter

import (
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

// XformOpt is a function that is used to configure the Transform func's behavior.
type XformOpt func(*xformCfg)

type xformCfg struct {
	// fields tracks both inclusion of fields, and what the output name should be.
	fields map[string]string
	// mungers potentially modify a row.  See the definition of the Munger type
	// for more information. They're not keyed to a specific field, because it's
	// expected that there will be a small number of them compared to fields, so
	// it's (maybe) cheaper to just iterate through all the mungers instead of
	// checking if every field has one.
	mungers []Munger
}

// Transform transforms a row based on the provided XformOpts.
func Transform(in types.TableContent, opts ...XformOpt) (types.TableContent, error) {
	var out []map[string]interface{}
	var err error

	var cfg xformCfg
	for _, f := range opts {
		f(&cfg)
	}

	for i := range in { // For every row in the input

		row := in[i]
		for _, f := range cfg.mungers {
			row, err = f(row)
			if err != nil {
				return nil, fmt.Errorf("error munging row %v: %v", i, err)
			}
		}

		if row == nil {
			continue
		}

		new := map[string]interface{}{} // create an empty output row.

		if len(cfg.fields) > 1 {
			// If a field map is specified, use it to rename and filter.
			// (There's always one entry for "id")
			for k, v := range row {
				if nk, ok := cfg.fields[k]; ok {
					new[nk] = v
				}
			}
		} else {
			// Otherwise, keep everything.
			new = row
		}
		out = append(out, new)
	}
	return out, err
}

// WithFieldMap configures the transformer to include specific fields and rename them.  Map keys are old: new.
func WithFieldMap(fields map[string]string) XformOpt {
	// "id" is always retained.
	fields["id"] = "id"

	return func(cfg *xformCfg) {
		cfg.fields = fields
	}
}

// WithFieldSlice configures the transformer to include specific fields where the input and output names are identical.
func WithFieldSlice(allowedKeys []string) XformOpt {
	keys := make(map[string]string, len(allowedKeys))
	for _, k := range allowedKeys {
		keys[k] = k
	}
	return WithFieldMap(keys)
}

// A Munger function accepts a single row, potentially copies it, modifies it,
// and returns the row.  Because a row is a pointer to shared structure, be sure
// to copy if you don't want to copy the input.
// munge, verb: to manipulate or transform data (https://www.google.com/search?q=define+munge)
type Munger func(in map[string]interface{}) (map[string]interface{}, error)

// WithMunger configures the transformer to include a specific munge function.
func WithMunger(m Munger) XformOpt {
	return func(cfg *xformCfg) {
		cfg.mungers = append(cfg.mungers, m)
	}
}
