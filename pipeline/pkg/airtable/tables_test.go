package airtable

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

// stubFetcher is a stub Fetcher for when the test only needs one table.
type stubFetcher struct {
	content []map[string]interface{}
	err     error
}

func (sf *stubFetcher) Download(_ context.Context, _ string) (types.TableContent, error) {
	return sf.content, sf.err
}

// stubMultiFetcher is a Fetcher for when the test needs multiple tables.
type stubMultiFetcher struct {
	content map[string][]map[string]interface{}
	err     error
}

func (sf *stubMultiFetcher) Download(_ context.Context, table string) (types.TableContent, error) {
	d, ok := sf.content[table]
	if !ok {
		return nil, fmt.Errorf("table %q data not specified in test", table)
	}
	return d, sf.err
}
func TestTables_GetCounties(t *testing.T) {
	f := &stubFetcher{content: []map[string]interface{}{
		{
			"id":   "recA",
			"name": "test county",
		},
	},
	}

	tables := NewFakeTables(context.Background(), f)

	// Why does this do this twice?
	for i := 0; i < 2; i++ {
		table, err := tables.GetCounties(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, table[0]["name"], "test county")
	}
}

func TestTables_GetProviders(t *testing.T) {
	f := &stubFetcher{
		content: []map[string]interface{}{
			{
				"id":   "recA",
				"name": "test provider",
			},
		},
	}

	tables := NewFakeTables(context.Background(), f)

	// Why does this do this twice?
	for i := 0; i < 2; i++ {
		table, err := tables.GetProviders(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, table[0]["name"], "test provider")
	}
}

func TestGetTables_XFormError(t *testing.T) {
	f := &stubFetcher{
		content: []map[string]interface{}{
			{},
		},
	}

	// returnError is a test munger that just returns an error (to test error
	// handling).  borrowed from xform_test.go
	returnError := func(row map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("fail")
	}

	ctx := context.Background()
	tables := NewFakeTables(ctx, f)

	_, err := tables.getTable(ctx, "providers", filter.WithMunger(returnError))
	assert.Error(t, err)
}

func TestGetTables_MultipleErr(t *testing.T) {
	f := &stubFetcher{
		content: []map[string]interface{}{
			{},
		},
		err: errors.New("Fetching"),
	}
	returnError := func(row map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("Munging")
	}

	ctx := context.Background()
	tables := NewFakeTables(ctx, f)

	_, err := tables.getTable(ctx, "providers", filter.WithMunger(returnError))
	assert.Error(t, err)
	// Failures in fetching should pre-empt failures in munging
	assert.Equal(t, "Fetching", err.Error())
}

func TestTables_CachedErr(t *testing.T) {
	f := &stubFetcher{
		content: []map[string]interface{}{
			{
				"name": "test county",
			},
		},
		err: errors.New("Failure"),
	}

	ctx := context.Background()
	tables := NewFakeTables(ctx, f)

	_, err := tables.GetCounties(ctx)
	assert.Error(t, err)

	// Should still fail, caching the err from last time, if called again, even if the underlying fetcher returns success.
	f.err = nil
	_, err = tables.GetCounties(ctx)
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
			desc: "remove if missing",
			in: map[string]interface{}{
				"Latest report notes": []string{"a", "b"},
			},
			want: map[string]interface{}{
				"Latest report notes": "",
			},
		},
		{
			desc: "remove if malformed",
			in: map[string]interface{}{
				"Latest report yes?":  "i'm not a number",
				"Latest report notes": []string{"a", "b"},
			},
			want: map[string]interface{}{
				"Latest report yes?":  "i'm not a number",
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
	f := &stubMultiFetcher{
		content: map[string][]map[string]interface{}{
			"Locations": {
				{
					"id":                  "1",
					"Name":                "yes=yes, all fields explicit",
					"Latest report yes?":  1.0,
					"Latest report notes": []string{"a", "b"},
					"is_soft_deleted":     false,
				},
				{
					"id":                  "2",
					"Name":                "yes=no, is_soft_deleted implicitly false",
					"Latest report yes?":  0.0,
					"Latest report notes": []string{"c", "d"},
				},
				{
					"id":                  "3",
					"Name":                "soft_deleted",
					"Latest report yes?":  0.0,
					"Latest report notes": []string{"c", "d"},
					"is_soft_deleted":     true,
				},
				{
					"id":                  "4",
					"Name":                "missing latest report yes? implicit no",
					"Latest report notes": []string{"c", "d"},
				},
				{
					"id":                                  "5",
					"Name":                                "this location uses county scheduling",
					"County":                              "Los Angeles County",
					"Appointment scheduling instructions": "Uses county scheduling system",
					"Latest report yes?":                  1.0,
					"Latest report notes":                 []string{"a", "b"},
					"is_soft_deleted":                     false,
				},
				{
					"id":                                  "6",
					"Name":                                "this location uses county scheduling but is in a county we don't know about",
					"County":                              "Imaginary County",
					"Appointment scheduling instructions": "Uses county scheduling system",
					"Latest report yes?":                  1.0,
					"Latest report notes":                 []string{"a", "b"},
					"is_soft_deleted":                     false,
				},
				{
					"id":                                  "7",
					"Name":                                "this location uses doesn't have a county specified",
					"Appointment scheduling instructions": "Uses county scheduling system",
					"Latest report yes?":                  1.0,
					"Latest report notes":                 []string{"a", "b"},
					"is_soft_deleted":                     false,
				},
				{
					"id": "8",
					// no other fields, will get dropped by dropEmpty.
					// if it's not dropped, the test will fail because it's not in the expected results.
				},
			},
			"Counties": {
				{
					"id":                                  "recA",
					"County":                              "Los Angeles County",
					"County vaccination reservations URL": "http://publichealth.lacounty.gov/acd/ncorona2019/vaccine/hcwsignup/",
				},
				{
					"id":    "recB",
					"weird": "This record has none of the normal County fields.  It gets ignroed.",
				},
				{
					"id":          "recC",
					"County":      12345,
					"Description": "This record has the wrong type for County.  It should be a string.",
				},
				{
					"id": "recZ",
					// no other fields, will get dropped by dropEmpty.
				},
			},
		},
	}

	want := types.TableContent{
		{
			"id":                  "1",
			"Name":                "yes=yes, all fields explicit",
			"Latest report yes?":  1.0,
			"Latest report notes": []string{"a", "b"},
			"is_soft_deleted":     false,
		},
		{
			"id":                  "2",
			"Name":                "yes=no, is_soft_deleted implicitly false",
			"Latest report yes?":  0.0,
			"Latest report notes": "",
		},
		{
			"id":                  "4",
			"Name":                "missing latest report yes? implicit no",
			"Latest report notes": "",
		},
		{
			"id":                                  "5",
			"Name":                                "this location uses county scheduling",
			"County":                              "Los Angeles County",
			"Appointment scheduling instructions": "http://publichealth.lacounty.gov/acd/ncorona2019/vaccine/hcwsignup/",
			"Latest report yes?":                  1.0,
			"Latest report notes":                 []string{"a", "b"},
			"is_soft_deleted":                     false,
		},
		{
			"id":                                  "6",
			"Name":                                "this location uses county scheduling but is in a county we don't know about",
			"County":                              "Imaginary County",
			"Appointment scheduling instructions": "Uses county scheduling system",
			"Latest report yes?":                  1.0,
			"Latest report notes":                 []string{"a", "b"},
			"is_soft_deleted":                     false,
		},
		{
			"id":                                  "7",
			"Name":                                "this location uses doesn't have a county specified",
			"Appointment scheduling instructions": "Uses county scheduling system",
			"Latest report yes?":                  1.0,
			"Latest report notes":                 []string{"a", "b"},
			"is_soft_deleted":                     false,
		},
	}

	ctx := context.Background()
	tables := NewFakeTables(ctx, f)

	got, err := tables.GetLocations(ctx)
	assert.NoError(t, err)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected locations: -want +got\n%v\n", diff)
	}
}

// TestGetLocations_NoCounties tests a corner case where retrieving the Counties
// table fails (because GetLocations is now dependent on GetCounties).  This
// isn't covered by TestGetLocations() above because it uses a single Locations
// and Counties table.
func TestGetLocations_NoCounties(t *testing.T) {
	f := &stubMultiFetcher{
		content: map[string][]map[string]interface{}{
			"Locations": {
				{
					"id":                  "1",
					"Name":                "yes=yes, all fields explicit",
					"Latest report yes?":  1.0,
					"Latest report notes": []string{"a", "b"},
					"is_soft_deleted":     false,
				},
			},
		},
	}

	ctx := context.Background()
	tables := NewFakeTables(ctx, f)

	_, err := tables.GetLocations(ctx)
	if err == nil {
		t.Errorf("want error, got nil")
	}
}
