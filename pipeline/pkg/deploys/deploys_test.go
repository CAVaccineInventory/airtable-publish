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
	}{
		"prod":    {envVar: "prod", deploy: DeployProduction},
		"staging": {envVar: "staging", deploy: DeployStaging},
		"testing": {envVar: "testing", deploy: DeployTesting},
		"blank":   {envVar: "", deploy: DeployTesting},
	}
	t.Cleanup(func() {
		os.Unsetenv("DEPLOY")
	})
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("DEPLOY", tc.envVar)
			bucket, err := GetUploadURL(LegacyVersion)
			require.NoError(t, err)
			if !strings.HasPrefix(bucket, "gs://") {
				t.Errorf("Upload URL does not start with gs://")
			}

			url, err := GetDownloadURL(LegacyVersion)
			require.NoError(t, err)
			if !strings.HasPrefix(url, "https://") {
				t.Errorf("Download URL does not start with https://")
			}
			if strings.HasSuffix(url, "/") {
				t.Errorf("Download URL ends with /")
			}
		})
	}
}
