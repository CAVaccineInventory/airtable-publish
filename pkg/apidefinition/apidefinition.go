package apidefinition

import "github.com/CAVaccineInventory/airtable-export/pkg/table"

type Definition struct {
	Group string
	Version int
	Stability string
	AllowedFields []string
	TransformFunc func(table.Table) (map[string]interface{}, error)
}
