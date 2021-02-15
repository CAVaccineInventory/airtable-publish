package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
	"github.com/stretchr/testify/require"
)

var seq = 0

// synthesizeIDs creates synthetic IDs for a table.  Sometimes the saved test data on disk doesn't have an ID stored.
func synthesizeIDs(c airtable.TableContent) {
	for _, r := range c {
		seq = seq + 1
		r["id"] = fmt.Sprintf("%08x", seq)
	}
}

func TestSanitize(t *testing.T) {
	tests := map[string]struct {
		endpointFunc endpointFunc
		testDataFile string
		badKeys      []string // fields no record must have
		requiredKeys []string // fields every record must have
	}{
		"Locations": {
			endpointFunc: EndpointMap[deploys.LegacyVersion]["Locations"],
			testDataFile: "test_data/locations_reduced.json",
			badKeys:      []string{"Last report author", "Internal notes"},
			requiredKeys: []string{"Name"},
		},
		"Counties": {
			endpointFunc: EndpointMap[deploys.LegacyVersion]["Counties"],
			testDataFile: "test_data/counties.json",
			badKeys:      []string{"Internal notes"},
		},
		"Locations-V1": {
			endpointFunc: EndpointMap[deploys.VersionType("1")]["locations"],
			testDataFile: "test_data/locations_reduced.json",
			badKeys:      []string{"Last report author", "Internal notes"},
			requiredKeys: []string{"Name"},
		},
		"Counties-V1": {
			endpointFunc: EndpointMap[deploys.VersionType("1")]["counties"],
			testDataFile: "test_data/counties.json",
			badKeys:      []string{"Internal notes"},
		},
		"Providers-V1": {
			endpointFunc: EndpointMap[deploys.VersionType("1")]["providers"],
			testDataFile: "test_data/providers.json",
			badKeys:      []string{"airtable_id"},
		},
	}

	ctx := context.Background()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// "id" is always required
			tc.requiredKeys = append(tc.requiredKeys, "id")

			getData := func(ctx context.Context, tableName string) (airtable.TableContent, error) {
				o, err := airtable.ObjectFromFile(ctx, name, tc.testDataFile)
				if err != nil {
					return nil, err
				}
				synthesizeIDs(o)
				return o, nil
			}
			fakeTables := airtable.NewFakeTables(getData)
			out, err := tc.endpointFunc(ctx, fakeTables)
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

			for i, l := range locs {
				// Check for bad keys.
				for _, k := range tc.badKeys {
					if _, ok := l[k]; ok {
						t.Errorf("bad key %q found in row %d", k, i)
					}
				}
				// Check for required keys.
				for _, k := range tc.requiredKeys {
					if _, ok := l[k]; !ok {
						t.Errorf("key %q m missing in row %d", k, i)
					}
				}

			}
		})
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
