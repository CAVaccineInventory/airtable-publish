package airtable

import (
	"context"
)

// NewFakeTables provides a fake implementation of Tables, to use in tests that expect tables.
func NewFakeTables(fetchFunc func(context.Context, string) (TableContent, error)) *Tables {
	tables := NewTables()
	tables.fetchFunc = fetchFunc
	return tables
}
