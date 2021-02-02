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
