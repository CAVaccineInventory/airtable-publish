package airtable

import (
	"context"
	"testing"
)

func TestObjectFromFile(t *testing.T) {
	tests := []struct {
		desc       string
		filename   string
		wantLength int
		wantErr    bool
	}{
		{
			desc:       "success",
			filename:   "test_data/two_counties.json",
			wantLength: 2,
		},
		{
			desc:     "file does not exist",
			filename: "test_data/doesnotexist.json",
			wantErr:  true,
		},
		{
			desc:     "not valid json",
			filename: "test_data/notvalidjson.json",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx := context.Background()
			o, err := ObjectFromFile(ctx, "counties", tt.filename)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
			if len(o) != tt.wantLength {
				t.Errorf("got %v records, want %v", len(o), tt.wantLength)
			}

		})
	}
}
