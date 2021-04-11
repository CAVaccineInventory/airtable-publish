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
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/stretchr/testify/require"
)

var seq = 0

// synthesizeIDs creates synthetic IDs for a table.  Sometimes the saved test data on disk doesn't have an ID stored.
func synthesizeIDs(c types.TableContent) {
	for _, r := range c {
		seq = seq + 1
		r["id"] = fmt.Sprintf("%08x", seq)
	}
}

type stubFetchFromFile struct {
	name, dataFile string
}

func (sf *stubFetchFromFile) Download(ctx context.Context, _ string) (types.TableContent, error) {
	o, err := airtable.ObjectFromFile(ctx, sf.name, sf.dataFile)
	if err != nil {
		return nil, err
	}
	synthesizeIDs(o)
	return o, nil
}

func TestSanitize(t *testing.T) {
	tests := map[string]struct {
		endpointFunc endpointFunc
		testDataFile string
		badKeys      []string // fields no record must have
		requiredKeys []string // fields every record must have
	}{
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

			f := &stubFetchFromFile{
				name:     name,
				dataFile: tc.testDataFile,
			}

			fakeTables := airtable.NewFakeTables(ctx, f)
			out, err := tc.endpointFunc(ctx, fakeTables)
			require.NoError(t, err)

			got, err := storage.Serialize(out)
			require.NoError(t, err)

			//  Basic sanity check
			if bytes.Contains(got.Bytes(), []byte("@gmail.com")) {
				t.Errorf("result contains @gmail.com")
			}

			locs := make(types.TableContent, 0)
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
		wantErr     bool
	}{
		"Providers": {
			deploy:      "prod",
			containsURL: "https://api.vaccinateca.com/v1/providers.json",
		},
		"Providers-baddeploy": {
			deploy:      "error",
			containsURL: "https://api.vaccinateca.com/v1/providers.json",
			wantErr:     true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("DEPLOY", tc.deploy)
			URLs, err := EndpointURLs()
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error from EndpointURLs(): %v", err)
			}
			if err == nil {
				require.Contains(t, URLs, tc.containsURL)
			}
		})
	}
}

func TestEndPointAccessors(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("DEPLOY")
	})

	e := Endpoint{
		Version:  "1",
		Resource: "locations",
		// Transform: Not specified, because we're not testing it
	}

	tests := []struct {
		desc       string
		deploy     string
		wantURL    string
		wantString string
		wantErr    bool
	}{
		{
			desc:       "success",
			deploy:     "prod",
			wantURL:    "https://api.vaccinateca.com/v1/locations.json",
			wantString: "1/locations",
			wantErr:    false,
		},
		{
			desc:       "error",
			deploy:     "doesnotexist",
			wantString: "1/locations",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			os.Setenv("DEPLOY", tt.deploy)

			gotString := e.String()
			if gotString != tt.wantString {
				t.Errorf("String(): got %q, want %q", gotString, tt.wantString)
			}

			gotURL, err := e.URL()
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}

			if gotURL != tt.wantURL {
				t.Errorf("String(): got %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}
