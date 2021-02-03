package filter

import (
	"reflect"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
)

func TestFilterToAllowedKeys(t *testing.T) {
	cases := []struct {
		input     airtable.TableContent
		allowKeys []string
		expect    airtable.TableContent
	}{
		{
			input: airtable.TableContent{
				{
					"allow":    "allowvalue",
					"disallow": "disallowvalue",
				},
			},
			allowKeys: []string{"allow"},
			expect: airtable.TableContent{
				{
					"allow": "allowvalue",
				},
			},
		},
	}

	for _, c := range cases {
		actual := ToAllowedKeys(c.input, c.allowKeys)
		if !reflect.DeepEqual(c.expect, actual) {
			t.Errorf("Expected all and only allowed keys.\nGOT: %v\nEXPECTED: %v\n", actual, c.expect)
		}
	}
}
