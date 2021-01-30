package endpoints

import "github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"

var AllEndpoints = generator.EndpointMap{
	"Locations": GenerateV1Locations,
	"Counties":  GenerateV1Counties,
}
