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

// Provide a configuration suitable for configuring Stackdriver
func StackdriverOptions(namespace string) stackdriver.Options {
	location, err := metadata.Zone()
	if err != nil {
		location = "unknown"
	}
	return stackdriver.Options{
		ProjectID:         GoogleProjectID,
		ReportingInterval: 60 * time.Second,
		Resource: &monitoredres.MonitoredResource{
			Type: "generic_node",
			Labels: map[string]string{
				"project_id": GoogleProjectID,
				"location":   location,
				"namespace":  namespace,
				"node_id":    getTaskID(),
			},
		},
	}
}
