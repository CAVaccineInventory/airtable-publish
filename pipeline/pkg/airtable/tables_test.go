package airtable

import (
	"context"
	"errors"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestTables_GetCounties(t *testing.T) {
	fetchFunc := func(_ context.Context, _ string) (types.TableContent, error) {
		return []map[string]interface{}{
			{
				"name": "test county",
			},
		}, nil
	}

	tables := NewTables()
	tables.fetchFunc = fetchFunc

	for i := 0; i < 2; i++ {
		table, err := tables.GetCounties(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, table[0]["name"], "test county")
	}
}

func TestTables_CachedErr(t *testing.T) {
	fail := true
	fetchFunc := func(_ context.Context, _ string) (types.TableContent, error) {
		if fail {
			return nil, errors.New("Failure")
		}
		return []map[string]interface{}{
			{
				"name": "test county",
			},
		}, nil
	}

	tables := NewTables()
	tables.fetchFunc = fetchFunc

	_, err := tables.GetCounties(context.Background())
	assert.Error(t, err)

	// Should still fail, caching the err from last time, if called again
	fail = false
	_, err = tables.GetCounties(context.Background())
	assert.Error(t, err)
}
