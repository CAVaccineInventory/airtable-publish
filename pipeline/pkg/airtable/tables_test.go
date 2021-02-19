package airtable

import (
	"context"
	"errors"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/google/go-cmp/cmp"
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

	tables := NewFakeTables(fetchFunc)

	for i := 0; i < 2; i++ {
		table, err := tables.GetCounties(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, table[0]["name"], "test county")
	}
}

func TestTables_GetProviders(t *testing.T) {
	fetchFunc := func(_ context.Context, _ string) (types.TableContent, error) {
		return []map[string]interface{}{
			{
				"name": "test provider",
			},
		}, nil
	}

	tables := NewFakeTables(fetchFunc)

	for i := 0; i < 2; i++ {
		table, err := tables.GetProviders(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, table[0]["name"], "test provider")
	}
}

func TestGetTables_XFormError(t *testing.T) {
	fetchFunc := func(_ context.Context, _ string) (types.TableContent, error) {
		return []map[string]interface{}{
			{},
		}, nil
	}

	// returnError is a test munger that just returns an error (to test error
	// handling).  borrowed from xform_test.go
	returnError := func(row map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("fail")
	}

	tables := NewFakeTables(fetchFunc)

	_, err := tables.getTable(context.Background(), "providers", filter.WithMunger(returnError))
	assert.Error(t, err)
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

	tables := NewFakeTables(fetchFunc)

	_, err := tables.GetCounties(context.Background())
	assert.Error(t, err)

	// Should still fail, caching the err from last time, if called again
	fail = false
	_, err = tables.GetCounties(context.Background())
	assert.Error(t, err)
}

func TestHideNotes(t *testing.T) {
	tests := []struct {
		desc string
		in   map[string]interface{}
		want map[string]interface{}
	}{
		{
			desc: "remove if not yes",
			in: map[string]interface{}{
				"Latest report yes?":  0.0,
				"Latest report notes": []string{"a", "b"},
			},
			want: map[string]interface{}{
				"Latest report yes?":  0.0,
				"Latest report notes": "",
			},
		},
		{
			desc: "don't remove if yes",
			in: map[string]interface{}{
				"Latest report yes?":  1.0,
				"Latest report notes": []string{"a", "b"},
			},
			want: map[string]interface{}{
				"Latest report yes?":  1.0,
				"Latest report notes": []string{"a", "b"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := hideNotes(tt.in)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("-want +got:\n%v\n", diff)
			}
		})
	}
}

func TestGetLocations(t *testing.T) {
	fetchFunc := func(_ context.Context, _ string) (types.TableContent, error) {
		return []map[string]interface{}{
			{
				"id":                  "1",
				"Name":                "test location 1",
				"Latest report yes?":  1.0,
				"Latest report notes": []string{"a", "b"},
				"is_soft_deleted":     false,
			},
			{
				"id":                  "2",
				"Name":                "test location 2",
				"Latest report yes?":  0.0,
				"Latest report notes": []string{"c", "d"},
			},
			{
				"id":                  "3",
				"Name":                "deleted location",
				"Latest report yes?":  0.0,
				"Latest report notes": []string{"c", "d"},
				"is_soft_deleted":     true,
			},
		}, nil
	}

	want := types.TableContent{
		{
			"id":                  "1",
			"Name":                "test location 1",
			"Latest report yes?":  1.0,
			"Latest report notes": []string{"a", "b"},
			"is_soft_deleted":     false,
		},
		{
			"id":                  "2",
			"Name":                "test location 2",
			"Latest report yes?":  0.0,
			"Latest report notes": "",
		},
	}

	tables := NewFakeTables(fetchFunc)

	got, err := tables.GetLocations(context.Background())
	assert.NoError(t, err)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected locations: -want +got\n%v\n", diff)
	}

}
