package airtable

import (
	"context"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

// NewFakeTables provides a fake implementation of Tables, to use in tests that expect tables.
func NewFakeTables(fetchFunc func(context.Context, string) (types.TableContent, error)) *Tables {
	tables := NewTables()
	tables.fetchFunc = fetchFunc
	return tables
}
