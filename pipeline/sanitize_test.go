package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitize(t *testing.T) {
	tests := map[string]struct {
		testDataFile string
		badKeys      []string
	}{
		"Locations": {testDataFile: "test_data/locations_reduced.json", badKeys: []string{"Last report author", "Internal notes"}},
		"Counties":  {testDataFile: "test_data/counties.json", badKeys: []string{"Internal notes"}},
	}

	for name, tc := range tests {
		in, err := ObjectFromFile(name, tc.testDataFile)
		require.NoError(t, err)

		got, err := Sanitize(in, name)
		require.NoError(t, err)

		//  Basic sanity check
		if bytes.Contains(got.Bytes(), []byte("@gmail.com")) {
			t.Errorf("result contains @gmail.com")
		}

		locs := make([]map[string]interface{}, 0)
		err = json.Unmarshal(got.Bytes(), &locs)
		require.NoError(t, err)

		// Check a sampling of bad keys.
		for _, l := range locs {
			for _, k := range tc.badKeys {
				if _, ok := l[k]; ok {
					t.Errorf("bad key %v found in ", k)
				}
			}
		}

	}
}
