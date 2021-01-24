package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSanitize(t *testing.T) {
	in, err := ObjectFromFile("test_data/locations_reduced.json")

	got, err := Sanitize(in)
	require.NoError(t, err)

	//  Basic sanity check
	if bytes.Contains(got.Bytes(), []byte("@gmail.com")) {
		t.Errorf("result contains @gmail.com")
	}

	locs := make([]map[string]interface{}, 0)
	err = json.Unmarshal(got.Bytes(), &locs)
	require.NoError(t, err)

	// Check a sampling of bad keys.  (Caution: This is a little fragile because
	// of case sensitivity.)
	badKeys := []string{
		"Last report author",
		"Internal notes",
	}

	for _, l := range locs {
		for _, k := range badKeys {
			if _, ok := l[k]; ok {
				t.Errorf("bad key %v found in ", k)
			}
		}
	}
}
