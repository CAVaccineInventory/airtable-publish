package filter

import (
	"strings"
	"testing"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/google/go-cmp/cmp"
)

// ucName is a test munger demonstrating changing the content of a field.
func ucName(row map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(row))

	for k, v := range row {
		nv := v.(string)
		if k == "Name" {
			nv = strings.ToUpper(nv)
		}
		out[k] = nv
	}
	return out, nil
}

// ucName is a test munger demonstrating changing the content of a field inplace.
func ucNameInplace(row map[string]interface{}) (map[string]interface{}, error) {
	row["Name"] = strings.ToUpper(row["Name"].(string))
	return row, nil
}

// dropNameInPlace is a test munger demonstrating dropping a field from a row inplace.
func dropNameInplace(row map[string]interface{}) (map[string]interface{}, error) {
	delete(row, "Name")
	return row, nil
}

// dropNameInPlace is a test munger demonstrating dropping a field from a row.
func dropName(row map[string]interface{}) (map[string]interface{}, error) {
	var out = make(map[string]interface{})
	for k, v := range row {
		if k != "Name" {
			out[k] = v
		}
	}
	return out, nil
}

// dropRow is a test munger demonstrating dropping a row entirely if it meets conditions.
func dropRow(row map[string]interface{}) (map[string]interface{}, error) {
	if row["Name"].(string) == "Moderna" {
		return nil, nil
	}
	return row, nil
}

// testData is a helper function to return the data we use for the tests in this file.  It exists only to reduce a little duplication.
func testData() types.TableContent {
	return types.TableContent{
		map[string]interface{}{
			"id":    "1",
			"Name":  "Moderna",
			"Other": "MiXeD",
		},
		map[string]interface{}{
			"id":    "2",
			"Name":  "Pfizer",
			"Other": "lower",
		},
	}
}

func TestTransformMunger(t *testing.T) {
	tests := []struct {
		desc    string
		in      types.TableContent
		want    types.TableContent
		munger  Munger
		inplace bool
	}{
		{
			desc: "inplace",
			in:   testData(),
			want: types.TableContent{
				map[string]interface{}{
					"id":    "1",
					"Name":  "MODERNA",
					"Other": "MiXeD",
				},
				map[string]interface{}{
					"id":    "2",
					"Name":  "PFIZER",
					"Other": "lower",
				},
			},
			inplace: true,
			munger:  ucNameInplace,
		},
		{
			desc: "not-inplace",
			in:   testData(),
			want: types.TableContent{
				map[string]interface{}{
					"id":    "1",
					"Name":  "MODERNA",
					"Other": "MiXeD",
				},
				map[string]interface{}{
					"id":    "2",
					"Name":  "PFIZER",
					"Other": "lower",
				},
			},
			inplace: false,
			munger:  ucName,
		},
		{
			desc: "delete field inplace",
			in:   testData(),
			want: types.TableContent{
				map[string]interface{}{
					"id":    "1",
					"Other": "MiXeD",
				},
				map[string]interface{}{
					"id":    "2",
					"Other": "lower",
				},
			},
			inplace: true,
			munger:  dropNameInplace,
		},
		{
			desc: "delete field",
			in:   testData(),
			want: types.TableContent{
				map[string]interface{}{
					"id":    "1",
					"Other": "MiXeD",
				},
				map[string]interface{}{
					"id":    "2",
					"Other": "lower",
				},
			},
			inplace: false,
			munger:  dropName,
		},
		{
			desc: "drop row",
			in:   testData(),
			want: types.TableContent{
				map[string]interface{}{
					"id":    "2",
					"Name":  "Pfizer",
					"Other": "lower",
				},
			},
			inplace: false,
			munger:  dropRow,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := Transform(tt.in, WithMunger(tt.munger))
			if err != nil {
				t.Fatalf("unexpected error from Transform: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("-want +got:\n %v\n", diff)
			}
			if tt.inplace {
				if diff := cmp.Diff(tt.in, got); diff != "" {
					t.Errorf("inplace edit, expected input edited: -want +got:\n %v\n", diff)
				}
			} else {
				if diff := cmp.Diff(testData(), tt.in); diff != "" {
					t.Errorf("expected input unmodified: -want +got:\n %v\n", diff)
				}
			}
		})
	}
}

func TestTransformFilter(t *testing.T) {
	tests := []struct {
		desc string
		in   types.TableContent
		want types.TableContent
		opts []XformOpt
	}{
		{
			desc: "no options specified - no transform",
			in:   testData(),
			want: testData(),
		},
		{
			desc: "no fields specified (slice) - keep all",
			in:   testData(),
			want: testData(),
			opts: []XformOpt{WithFieldSlice([]string{})},
		},
		{
			desc: "no fields specified (map) - keep all",
			in:   testData(),
			want: testData(),
			opts: []XformOpt{WithFieldMap(map[string]string{})},
		},
		{
			desc: "keep one (slice)",
			in:   testData(),
			want: types.TableContent{
				map[string]interface{}{
					"id":   "1",
					"Name": "Moderna",
				},
				map[string]interface{}{
					"id":   "2",
					"Name": "Pfizer",
				},
			},
			opts: []XformOpt{WithFieldSlice([]string{"Name"})},
		},
		{
			desc: "remap",
			in:   testData(),
			want: types.TableContent{
				map[string]interface{}{
					"id":     "1",
					"Nombre": "Moderna",
				},
				map[string]interface{}{
					"id":     "2",
					"Nombre": "Pfizer",
				},
			},
			opts: []XformOpt{WithFieldMap(map[string]string{"Name": "Nombre"})},
		},
		{
			desc: "id is implicit",
			in: types.TableContent{
				map[string]interface{}{
					"Name":  "AstraZenica",
					"id":    "kept even if not specified",
					"Color": "Green",
				},
			},
			want: types.TableContent{
				map[string]interface{}{
					"Name": "AstraZenica",
					"id":   "kept even if not specified",
				},
			},
			opts: []XformOpt{WithFieldSlice([]string{"Name"})},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := Transform(tt.in, tt.opts...)
			if err != nil {
				t.Fatalf("unexpected error from Transform: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("-want +got:\n %v\n", diff)
			}
		})
	}
}
