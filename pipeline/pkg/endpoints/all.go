package endpoints

import (
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/legacy"
)

// EndpointMap is a map of an API version, to all endpoints in that version.
//
// You can add new endpoints to a version, add fields to an endpoint
// in that version, tweak endpoint generation, etc.  IF YOU
// REMOVE/RENAME A FIELD OR CHANGE PROGRAMMATIC SEMANTICS, YOU NEED TO
// CREATE A NEW API VERSION.

// When creating an new API version, copy ALL endpoint definitions
// that should still exist into the new version, even if you are not
// changing them.  Reusing the original generate func between
// versions, if that endpoint is unchanged.  E.G. if changing the
// locations API in a breaking fashion, keep using the
// GenerateV1Counties func for v2, v3, etc.

var EndpointMap = map[deploys.VersionType]map[string]endpointFunc{
	deploys.LegacyVersion: {
		"Locations": legacy.Locations,
		"Counties":  legacy.Counties,
	},
}
