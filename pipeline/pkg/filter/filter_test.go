package filter

import (
	"errors"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

func TestCheckFields(t *testing.T) {
	cases := []struct {
		desc    string
		input   types.TableContent
		fields  []string
		wantErr error
	}{
		{
			desc: "no errors",
			input: types.TableContent{
				{
					"allow":    "allowvalue",
					"disallow": "disallowvalue",
				},
			},
			fields:  []string{"allow"},
			wantErr: nil,
		},
		{
			desc: "no errors - sparse fields",
			input: types.TableContent{
				{
					"a":        "a",
					"disallow": "disallowvalue",
				},
				{
					"b":        "b",
					"disallow": "disallowvalue",
				},
			},
			fields:  []string{"a", "b"},
			wantErr: nil,
		},
		{
			desc: "no errors - dense fields",
			input: types.TableContent{
				{
					"a":        "a",
					"b":        "b",
					"disallow": "disallowvalue",
				},
				{
					"a":        "a",
					"b":        "b",
					"disallow": "disallowvalue",
				},
			},
			fields:  []string{"a", "b"},
			wantErr: nil,
		},
		{
			desc: "missing field",
			input: types.TableContent{
				{
					"allow":    "allowvalue",
					"disallow": "disallowvalue",
				},
			},
			fields:  []string{"allow", "want"},
			wantErr: ErrMissingField,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			fieldMap := make(map[string]string, len(c.fields))
			for _, f := range c.fields {
				fieldMap[f] = f
			}
			err := checkFields(c.input, fieldMap)
			if !errors.Is(err, c.wantErr) {
				t.Errorf("got error %v, want %v", err, c.wantErr)
			}
		})
	}
}
