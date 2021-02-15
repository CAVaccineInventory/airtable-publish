package secrets

import (
	"context"
	"fmt"
	"os"
	"sync"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/config"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type SecretData struct {
	envName           string
	secretManagerName string
	lock              sync.Mutex
	value             string
	err               error
}

var (
	// AirtableSecret is used to fetch the raw data from Airtable; required
	AirtableSecret = &SecretData{
		envName:           "AIRTABLE_KEY",
		secretManagerName: "airtable-key",
		lock:              sync.Mutex{},
	}
	// HoneycombSecret is used to send metrics and spans to Honeycomb
	HoneycombSecret = &SecretData{
		envName:           "HONEYCOMB_KEY",
		secretManagerName: "honeycomb-key",
		lock:              sync.Mutex{},
	}
)

// Caches the secret for the lifetime of the process.
func Get(ctx context.Context, secret *SecretData) (string, error) {
	secret.lock.Lock()
	defer secret.lock.Unlock()

	if secret.value != "" || secret.err != nil {
		return secret.value, secret.err
	}

	envVal := os.Getenv(secret.envName)
	if envVal != "" {
		secret.value = envVal
		return envVal, nil
	}

	projectID, err := config.GetProject(ctx)
	if err != nil {
		secret.err = err
		return "", err
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		secret.err = fmt.Errorf("failed to make secretmanager client: %w", err)
		return "", secret.err
	}
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secret.secretManagerName),
	}
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		secret.err = err
		return "", err
	}

	secret.value = string(result.Payload.Data)
	return secret.value, nil
}

// Exits with code 1 if the airtable secret is not available.
func RequireAirtableSecret() {
	_, err := Get(context.Background(), AirtableSecret)
	if err != nil {
		fmt.Printf("No airtable secret could be fetched: %s\n", err)
		fmt.Println("Will not be able to read data, aborting!")
		fmt.Println("Set AIRTABLE_KEY or adjust rights in IAM")
		os.Exit(1)
	}
}
