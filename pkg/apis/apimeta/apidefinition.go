package apimeta

import (
	"github.com/CAVaccineInventory/airtable-export/pkg/table"
)

// EndpointDefinition is the definition of an API endpoint (e.g. /foo/v2/bar or /counties/v1/counties)
type EndpointDefinition struct {
	// Group is the name of the API group. A group is 1 or more Kinds that are versioned together.
	// This implies that the Kinds share schema or are directly related to one another.
	Group string
	// A kind is a type of resource (e.g. counties, locations). For now (TM) we only handle lists, so the kind is plural.
	Kind string
	// GenerateResponse takes a map of tables, and returns the API response for this API endpoint.
	GenerateResponse func(map[string]table.Table) (List, error)
}
