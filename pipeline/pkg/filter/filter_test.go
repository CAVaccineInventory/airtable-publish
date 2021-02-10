package filter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
)

func TestFilterRemapToAllowedKeys(t *testing.T) {
	cases := []struct {
		input   airtable.TableContent
		mapping map[string]string
		expect  airtable.TableContent
	}{
		{
			input: airtable.TableContent{
				{
					"allow":    "allowvalue",
					"disallow": "disallowvalue",
				},
			},
			mapping: map[string]string{"allow": "keep"},
			expect: airtable.TableContent{
				{
					"keep": "allowvalue",
				},
			},
		},
	}

	for _, c := range cases {
		actual, err := RemapToAllowedKeys(c.input, c.mapping)
		if err != nil {
			t.Fatalf("RemapToAllowedKeys: got error %q, want nil", err)
		}
		if !reflect.DeepEqual(c.expect, actual) {
			t.Errorf("Expected all and only allowed keys.\nGOT: %v\nEXPECTED: %v\n", actual, c.expect)
		}
	}
}

func TestCheckFields(t *testing.T) {
	cases := []struct {
		desc    string
		input   airtable.TableContent
		fields  []string
		wantErr error
	}{
		{
			desc: "no errors",
			input: airtable.TableContent{
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
			input: airtable.TableContent{
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
			input: airtable.TableContent{
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
			input: airtable.TableContent{
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
