package deploys

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestDeploys(t *testing.T) {
	tests := map[string]struct {
		wantError bool
		deploy    DeployType
	}{
		"prod":    {deploy: DeployProduction},
		"staging": {deploy: DeployStaging},
		"testing": {deploy: DeployTesting},
		"":        {deploy: DeployTesting},
		"bogus":   {deploy: DeployUnknown, wantError: true},
	}
	t.Cleanup(func() {
		os.Unsetenv("DEPLOY")
	})
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("DEPLOY", name)
			deploy, err := GetDeploy()
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, deploy, tc.deploy)
		})
	}
}

func TestDeployBuckets(t *testing.T) {
	tests := map[string]struct {
		envVar        string
		deploy        DeployType
		testingBucket string
		version       VersionType
		wantErr       bool
	}{
		"prod-legacy":    {envVar: "prod", deploy: DeployProduction, version: LegacyVersion},
		"staging-legacy": {envVar: "staging", deploy: DeployStaging, version: LegacyVersion},
		"testing-legacy": {envVar: "testing", deploy: DeployTesting, version: LegacyVersion},
		"blank-legacy":   {envVar: "", deploy: DeployTesting, version: LegacyVersion},
		"prod-v1":        {envVar: "prod", deploy: DeployProduction, version: "1"},
		"staging-v1":     {envVar: "staging", deploy: DeployStaging, version: "1"},
		"testing-v1":     {envVar: "testing", deploy: DeployTesting, version: "1"},
		"blank-v1":       {envVar: "", deploy: DeployTesting, version: "1"},
		"error-v1":       {envVar: "error", deploy: DeployTesting, version: "1", wantErr: true},
	}
	t.Cleanup(func() {
		os.Unsetenv("DEPLOY")
	})
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("DEPLOY", tc.envVar)
			bucket, err := GetUploadURL(tc.version)
			if (err != nil) != tc.wantErr {
				t.Errorf("unxpected error from GetUploadURL: %v", err)
			}
			if err == nil && !strings.HasPrefix(bucket, "gs://") {
				t.Errorf("Upload URL does not start with gs://")
			}

			url, err := GetDownloadURL(tc.version)
			if (err != nil) != tc.wantErr {
				t.Errorf("unxpected error from GetDownloadURL: %v", err)
			}
			if err == nil {
				if !strings.HasPrefix(url, "https://") {
					t.Errorf("Download URL does not start with https://")
				}
				if strings.HasSuffix(url, "/") {
					t.Errorf("Download URL ends with /")
				}
			}
		})
	}
}

func TestGetStorage(t *testing.T) {
	tests := []struct {
		desc    string
		deploy  string
		want    StorageWriter
		wantErr bool
	}{
		{
			desc:    "get testing storage",
			deploy:  string(DeployTesting),
			want:    storage.StoreLocal,
			wantErr: false,
		},
		{
			desc:    "non-existent deploy",
			deploy:  "made up string",
			wantErr: true,
		},
	}

	origDeploy := os.Getenv("DEPLOY")
	t.Cleanup(func() { os.Setenv("DEPLOY", origDeploy) })

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			os.Setenv("DEPLOY", tt.deploy)
			got, err := GetStorage()
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}

			// Checking that two functions are the same function is messy.
			gotV := reflect.ValueOf(got)
			wantV := reflect.ValueOf(tt.want)

			if !cmp.Equal(gotV.Pointer(), wantV.Pointer()) {
				t.Errorf("got %v, want %v", got, tt.want)
			}

		})
	}
}

func TestSetTestingStorage(t *testing.T) {
	orig := deploys[DeployTesting]
	t.Cleanup(func() { deploys[DeployTesting] = orig })

	const tbn = "testbucketname"
	SetTestingStorage(nil, "testbucketname")
	// Use nil for teseting the StorageWriter, because checking for an actual
	// function is a hassle.
	if deploys[DeployTesting].Storage != nil {
		t.Errorf("Storage: got %v, want %v", deploys[DeployTesting].Storage, nil)
	}
	if deploys[DeployTesting].LegacyBucket.Name != tbn {
		t.Errorf("LegacyBucket.Name: got %v, want %v", deploys[DeployTesting].LegacyBucket.Name, tbn)
	}
	if deploys[DeployTesting].APIBucket.Name != tbn {
		t.Errorf("APIBucket.Name: got %v, want %v", deploys[DeployTesting].APIBucket.Name, tbn)
	}
}
