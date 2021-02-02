package endpoints

import (
	"context"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
)

type endpointFunc func(context.Context, *airtable.Tables) (airtable.TableContent, error)

var EndpointMap = map[string]endpointFunc{
	"Locations": GenerateV1Locations,
	"Counties":  GenerateV1Counties,
}

type Endpoint struct {
	Resource  string
	Transform endpointFunc
}

func (ep *Endpoint) String() string {
	return ep.Resource
}

func AllEndpoints() []Endpoint {
	endpoints := make([]Endpoint, len(EndpointMap))
	i := 0
	for resource, transform := range EndpointMap {
		endpoints[i] = Endpoint{
			Resource:  resource,
			Transform: transform,
		}
		i++
	}
	return endpoints
}
