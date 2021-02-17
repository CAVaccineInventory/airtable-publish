package config

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

func Test_StackdriverOptions(t *testing.T) {
	// Make sure the "failure" case (not running in Cloud Functions returns the expected values.)

	origHTTPClient := httpClient
	t.Cleanup(func() {
		httpClient = origHTTPClient
	})
	httpClient = &http.Client{
		// Requests to the metadata server are expected to fail in this test
		// environments, so set an absurdly low timeout so they fail really
		// fast.  (If we're running in an environment with a reachable metadata
		// server, this may break the test, in which case we should make a
		// custom AlwaysImmediatelyFail RoundTripper.)
		Timeout: 1 * time.Millisecond,
	}

	ctx := context.Background()
	got := StackdriverOptions(ctx, "namespace1")
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

type stubMetadataClient struct {
	instanceID, zone string
}

func (c *stubMetadataClient) Zone() (string, error) {
	return c.zone, nil
}
func (c *stubMetadataClient) InstanceID() (string, error) {
	return c.instanceID, nil
}

func TestGetInstanceID(t *testing.T) {
	mc := &stubMetadataClient{
		zone:       "zone",
		instanceID: "abcdefghijklmnopqrstuvwxyz",
	}
	got, _ := getInstanceID(mc)
	if diff := cmp.Diff("stuvwxyz", got); diff != "" {
		t.Errorf("getInstanceID: -want +got:\n%v\n", diff)
	}
}
