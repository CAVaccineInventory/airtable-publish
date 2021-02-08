package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	traceapi "cloud.google.com/go/trace/apiv2"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"golang.org/x/oauth2/google"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

var (
	// This can be overridden in tests to set timeouts.
	httpClient *http.Client = nil
)

func getProject() (string, error) {

	// There's no obvious context to use here, so create one.  It would be
	// possible to plumb one through, but there's no obvious context available
	// to plumb.
	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()

	// Attempt to get project from local credentials.  This will work both on
	// Cloud or in a local machine.  On Cloud, we could use the metadata server,
	// but it's simpler to only implement one way.
	creds, err := google.FindDefaultCredentials(ctx, traceapi.DefaultAuthScopes()...)
	if err != nil {
		return "", fmt.Errorf("FindDefaultCredentials: %v", err)
	}
	if creds.ProjectID == "" {
		return "", errors.New("FindDefaultCredentials: no project found with application default credentials")
	}
	return creds.ProjectID, nil
}

func getInstanceID(mc *metadata.Client) (string, error) {
	inst, err := mc.InstanceID()
	if err != nil {
		return "", fmt.Errorf("metadata.InstanceID(): %w", err)
	}
	// InstanceIDs are really long, just pick the right-most characters.
	if len(inst) > 8 {
		inst = inst[len(inst)-8:]
	}
	return inst, nil
}

// Taken from https://github.com/census-ecosystem/opencensus-go-exporter-stackdriver/blob/master/stats.go#L182
func getTaskID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return fmt.Sprintf("go-%d@%s", os.Getpid(), hostname)
}

// StackdriverOptions provides a configuration suitable for configuring Stackdriver
func StackdriverOptions(namespace string) stackdriver.Options {
	mc := metadata.NewClient(httpClient)

	id, err := getInstanceID(mc)
	if err != nil {
		id = getTaskID()
	}

	location, err := mc.Zone()
	if err != nil {
		location = "unknown"
	}

	prj, err := getProject()
	if err != nil {
		prj = "unknown"
	}

	so := stackdriver.Options{
		ProjectID:         prj,
		ReportingInterval: 60 * time.Second,
		Resource: &monitoredres.MonitoredResource{
			Type: "generic_node",
			Labels: map[string]string{
				"project_id": prj,
				"location":   location,
				"namespace":  namespace,
				"node_id":    id,
			},
		},
	}
	log.Printf("StackDriver Labels: %+v", so.Resource.Labels) //nolint
	return so
}
