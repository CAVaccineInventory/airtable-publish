package endpoints

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
)

type endpointFunc func(context.Context, *airtable.Tables) (airtable.TableContent, error)

type Endpoint struct {
	Version   deploys.VersionType
	Resource  string
	Transform endpointFunc
}

func (ep *Endpoint) String() string {
	return fmt.Sprintf("%s/%s", ep.Version, ep.Resource)
}

func (ep *Endpoint) URL() (string, error) {
	baseURL, err := deploys.GetDownloadURL(ep.Version)
	if err != nil {
		return "", err
	}
	return baseURL + "/" + ep.Resource + ".json", nil
}

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
