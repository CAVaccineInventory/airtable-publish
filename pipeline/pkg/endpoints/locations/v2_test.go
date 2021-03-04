package locations

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_locationsTransformer(t *testing.T) {
	cases := []struct {
		name   string
		input  map[string]interface{}
		expect map[string]interface{}
	}{
		{
			name:   "field missing",
			input:  map[string]interface{}{"id": "test"},
			expect: map[string]interface{}{"id": "test"},
		},
		{
			name:   "appointment_scheduling_instructions transformed",
			input:  map[string]interface{}{"id": "test", "appointment_scheduling_instructions": []string{"example.com"}},
			expect: map[string]interface{}{"id": "test", "appointment_scheduling_instructions": "example.com"},
		},
		{
			name:   "appointment_scheduling_instructions has unexpected type",
			input:  map[string]interface{}{"id": "test", "appointment_scheduling_instructions": "example.com"},
			expect: map[string]interface{}{"id": "test"},
		},
		{
			name:   "appointment_scheduling_instructions has empty slice",
			input:  map[string]interface{}{"id": "test", "appointment_scheduling_instructions": []string{}},
			expect: map[string]interface{}{"id": "test"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual, err := locationsTransformer(c.input)
			assert.NoError(t, err)
			if !reflect.DeepEqual(c.expect, actual) {
				t.Errorf("Expect: %v\nGot: %v", c.expect, actual)
			}
		})
	}
}
