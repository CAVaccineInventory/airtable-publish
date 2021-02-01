package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
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
		out, err := EndpointMap[deploys.LegacyVersion][name](ctx, fakeTables)
		require.NoError(t, err)

		got, err := storage.Serialize(out)
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

func TestEndpoints(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("DEPLOY")
	})
	tests := map[string]struct {
		deploy      string
		containsURL string
	}{
		"Locations": {
			deploy:      "prod",
			containsURL: "https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Locations.json",
		},
		"Counties": {
			deploy:      "prod",
			containsURL: "https://storage.googleapis.com/cavaccineinventory-sitedata/airtable-sync/Counties.json",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("DEPLOY", tc.deploy)
			URLs, err := EndpointURLs()
			require.NoError(t, err)

			require.Contains(t, URLs, tc.containsURL)
		})
	}
}
