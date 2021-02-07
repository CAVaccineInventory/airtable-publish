package config

import (
	"fmt"
	"testing"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

func Test_StackdriverOptions(t *testing.T) {
	// Make sure the "failure" case (not running in Cloud Functions returns the expected values.)

	// This test is slow (5 seconds) because the http.Client used in the
	// metadata library has a non-overridable 750ms header response timeout.
	//
	// There is a non-Google implementation
	// (gitee.com/wangHvip/google-cloud-go/compute/metadata) that allows setting
	// a custom http.Client, so we could have a lower timeout for tests. Another
	// option would be to use package level variables, i.e. `metadataInstanceID
	// = metadata.InstanceID` which can be overridden with stubs in the tests.
	// Under the hood, the metadata library is just making simple HTTP calls
	// (and caching), so we could even just re-implement.
	//
	// This test may potentially fail on Cloud Build, if the build environment
	// metadata server is reachable.

	got := StackdriverOptions("namespace1")
	want := stackdriver.Options{
		ProjectID:         "unknown",
		ReportingInterval: 60 * time.Second,
		Resource: &monitoredres.MonitoredResource{
			Type: "generic_node",
			Labels: map[string]string{
				"project_id": "unknown",
				"location":   "unknown",
				"namespace":  "namespace1",
				"node_id":    getTaskID(),
			},
		},
	}

	// Compare the string representations. cmp.AllowUnexported and
	// cmpopts.IgnoreUnexported only look one level down, which makes for
	// complications when the unexported type is nested.
	gotS := fmt.Sprintf("%+v", got)
	wantS := fmt.Sprintf("%+v", want)
	if diff := cmp.Diff(wantS, gotS); diff != "" {
		t.Errorf("Unexpected results for StackdriverOptions(): -want +got\n: %s\n", diff)
	}
}
