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
