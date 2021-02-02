package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
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

	ctx := context.Background()
	for name, tc := range tests {
		getData := func(ctx context.Context, tableName string) (airtable.TableContent, error) {
			return airtable.ObjectFromFile(ctx, name, tc.testDataFile)
		}

		fakeTables := airtable.NewFakeTables(getData)
		out, err := EndpointMap[name](ctx, fakeTables)
		require.NoError(t, err)

		got, err := out.Serialize()
		require.NoError(t, err)

		//  Basic sanity check
		if bytes.Contains(got.Bytes(), []byte("@gmail.com")) {
			t.Errorf("result contains @gmail.com")
		}

		locs := make(airtable.TableContent, 0)
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
