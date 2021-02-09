package config

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

var (
	// This can be overridden in tests to set timeouts.
	httpClient *http.Client = nil
)

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

	prj, err := GetProject()
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
