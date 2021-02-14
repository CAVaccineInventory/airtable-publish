package endpoints

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

type endpointFunc func(context.Context, *airtable.Tables) (types.TableContent, error)

// Endpoint stores the data transform for a given version and resource
// path.
type Endpoint struct {
	Version   deploys.VersionType
	Resource  string
	Transform endpointFunc
}

func (ep *Endpoint) String() string {
	return fmt.Sprintf("%s/%s", ep.Version, ep.Resource)
}

// The download URL of the endpoint, based on the deploy, version, and resource.
func (ep *Endpoint) URL() (string, error) {
	baseURL, err := deploys.GetDownloadURL(ep.Version)
	if err != nil {
		return "", err
	}
	return baseURL + "/" + ep.Resource + ".json", nil
}

// Returns a list of Endpoints, based flattening the EndpointMap.
func AllEndpoints() []Endpoint {
	totalSize := 0
	for _, versionResources := range EndpointMap {
		totalSize += len(versionResources)
	}

	endpoints := make([]Endpoint, totalSize)
	i := 0
	for version, versionResources := range EndpointMap {
		for resource, transform := range versionResources {
			endpoints[i] = Endpoint{
				Version:   version,
				Resource:  resource,
				Transform: transform,
			}
			i++
		}
	}
	return endpoints
}

// Maps the URL() method over AllEndpoints.
func EndpointURLs() ([]string, error) {
	eps := AllEndpoints()
	endpointURLs := make([]string, len(eps))
	for i, ep := range eps {
		url, err := ep.URL()
		if err != nil {
			return nil, err
		}
		endpointURLs[i] = url
	}
	return endpointURLs, nil
}
