package config

import (
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

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

	id, err := metadata.InstanceID()
	if err != nil {
		id = getTaskID()
	}

	location, err := metadata.Zone()
	if err != nil {
		location = "unknown"
	}

	prj, err := metadata.ProjectID()
	if err != nil {
		prj = "unknown"
	}

	return stackdriver.Options{
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
}
