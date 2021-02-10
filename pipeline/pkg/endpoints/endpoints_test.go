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
		endpointFunc       endpointFunc
		testDataFile       string
		badKeys            []string
		expectRequiredKeys []string // Keys that should be in every element.
	}{
		"Locations": {
			endpointFunc:       EndpointMap[deploys.LegacyVersion]["Locations"],
			testDataFile:       "test_data/locations_reduced.json",
			badKeys:            []string{"Last report author", "Internal notes"},
			expectRequiredKeys: []string{"id"},
		},
		"Counties": {
			endpointFunc:       EndpointMap[deploys.LegacyVersion]["Counties"],
			testDataFile:       "test_data/counties.json",
			badKeys:            []string{"Internal notes"},
			expectRequiredKeys: []string{"id"},
		},
		"Locations-V1": {
			endpointFunc:       EndpointMap[deploys.VersionType("1")]["locations"],
			testDataFile:       "test_data/locations_reduced.json",
			badKeys:            []string{"Last report author", "Internal notes"},
			expectRequiredKeys: []string{"id"},
		},
		"Counties-V1": {
			endpointFunc:       EndpointMap[deploys.VersionType("1")]["counties"],
			testDataFile:       "test_data/counties.json",
			badKeys:            []string{"Internal notes"},
			expectRequiredKeys: []string{"id"},
		},
		"Providers-V1": {
			endpointFunc:       EndpointMap[deploys.VersionType("1")]["providers"],
			testDataFile:       "test_data/providers.json",
			expectRequiredKeys: []string{"id"},
		},
	}

	ctx := context.Background()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			getData := func(ctx context.Context, tableName string) (airtable.TableContent, error) {
				return airtable.ObjectFromFile(ctx, name, tc.testDataFile)
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

			elements := make(airtable.TableContent, 0)
			err = json.Unmarshal(got.Bytes(), &elements)
			require.NoError(t, err)

			for _, elem := range elements {
				// Check that no bad keys are present.
				for _, k := range tc.badKeys {
					if _, ok := elem[k]; ok {
						t.Errorf("bad key %s found in %v", k, elem)
					}
				}

				for _, k := range tc.expectRequiredKeys {
					if _, ok := elem[k]; !ok {
						t.Errorf("required key %s missing from %v", k, elem)
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
