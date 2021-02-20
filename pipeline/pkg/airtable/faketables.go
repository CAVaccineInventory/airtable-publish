package airtable

import (
	"context"
)

// NewFakeTables provides a fake implementation of Tables, to use in tests that expect tables.
func NewFakeTables(ctx context.Context, f fetcher) *Tables {
	t := NewTables("secret")
	t.fetcher = f
	return t
}
