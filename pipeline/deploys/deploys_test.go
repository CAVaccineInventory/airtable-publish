package deploys

import (
	"os"
	"strings"
	"testing"

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
		wantError     bool
		deploy        DeployType
		testingBucket string
	}{
		"prod":           {envVar: "prod", deploy: DeployProduction},
		"staging":        {envVar: "staging", deploy: DeployStaging},
		"testing":        {envVar: "testing", deploy: DeployTesting, wantError: true},
		"testing_bucket": {envVar: "testing", deploy: DeployTesting, testingBucket: "test-bucket-name"},
		"blank":          {envVar: "", deploy: DeployTesting, wantError: true},
		"blank_bucket":   {envVar: "", deploy: DeployTesting, testingBucket: "test-bucket-name"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("DEPLOY", tc.envVar)
			os.Setenv("TESTING_BUCKET", tc.testingBucket)
			bucket, err := GetExportBucket()
			if tc.wantError {
				require.Error(t, err)
				require.Equal(t, bucket, "")
			} else {
				require.NoError(t, err)
				if !strings.HasPrefix(bucket, "gs://") {
					t.Errorf("Bucket does not start with gs://")
				}
			}

			url, err := GetExportBaseURL()
			if tc.wantError {
				require.Error(t, err)
				require.Equal(t, url, "")
			} else {
				require.NoError(t, err)
				if !strings.HasPrefix(url, "https://") {
					t.Errorf("Bucket does not start with https://")
				}
				if strings.HasSuffix(url, "/") {
					t.Errorf("Bucket ends with /")
				}
			}
		})
	}
}
