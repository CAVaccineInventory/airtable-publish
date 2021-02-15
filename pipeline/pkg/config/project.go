package config

import (
	"context"
	"errors"
	"fmt"

	traceapi "cloud.google.com/go/trace/apiv2"
	"golang.org/x/oauth2/google"
)

// Returns the Google Cloud project identifier associated with the
// credentials that the application is running as.
func GetProject(ctx context.Context) (string, error) {
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
