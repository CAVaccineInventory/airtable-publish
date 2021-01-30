package airtable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTables_GetCounties(t *testing.T) {
	fetchFunc := func(_ context.Context, _ string) (TableContent, error) {
		return []map[string]interface{}{
			{
				"name": "test county",
			},
		}, nil
	}

	tables := NewTables()
	tables.fetchFunc = fetchFunc

	for i := 0; i < 2; i++ {
		table, err := tables.GetTable(context.Background(), "Counties")
		assert.NoError(t, err)
		assert.Equal(t, table[0]["name"], "test county")
	}
}
