package locations

import (
	"github.com/CAVaccineInventory/airtable-export/pkg/apidefinition"
	metav1 "github.com/CAVaccineInventory/airtable-export/pkg/apis/metadata/v1"
	"github.com/CAVaccineInventory/airtable-export/pkg/table"
)

type Object struct {
	Metadata metav1.Metadata
	Content  []map[string]interface{}
}

var LocationsV1 = apidefinition.Definition{
	Group: group,
	Version: 1,
	Stability: "unstable",
	// TODO: maybe TransFormfunc should be more freeform. I was originally imagining that table goes in, transofrmation happens, then metadata and HTML sanitization happens after.
	// If this is to match a strict signature, it would need to accept a set of tables in the future. An API may well join tables.
	TransformFunc: func(table table.Table) (Object, error) {
		return Object{
			Metadata: metav1.Metadata{
				ApiVersion: metav1.ApiVersion{
					Major: 1,
					Minor: 0,
					Stability: "unstable",
				},
				Contact: metav1.DefaultContact,
				UsageNotice: "TODO",
			},
			Content: todo,
		}, nil
	},
}
