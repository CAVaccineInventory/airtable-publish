package metadata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

func TestWrap(t *testing.T) {
	cases := []struct {
		name    string
		version VersionType
		expect  string
	}{
		{
			name:    "legacy",
			version: LegacyVersion,
			expect:  "[]",
		},
		{
			name:    "v1",
			version: "1",
			expect:  "{\"usage\":{\"contact\":{\"partnersEmail\":\"api@vaccinateca.com\"},\"documentation\":\"https://docs.vaccinateca.com\",\"notice\":\"Please contact VaccinateCA and let us know if you plan to rely on or publish this data. This data is provided with best-effort accuracy. If you are displaying this data, we expect you to display it responsibly. Please do not display it in a way that is easy to misread.\"},\"content\":[]}",
		},
		{
			name:    "v2",
			version: "2",
			expect:  "{\"metadata\":{\"usage\":{\"contact\":{\"partners_email\":\"api@vaccinateca.com\"},\"documentation\":\"https://docs.vaccinateca.com\",\"notice\":\"Please contact VaccinateCA and let us know if you plan to rely on or publish this data. This data is provided with best-effort accuracy. If you are displaying this data, we expect you to display it responsibly. Please do not display it in a way that is easy to misread.\"}},\"content\":[]}",
		},
	}

	tableContent := types.TableContent{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wrapped, err := Wrap(tableContent, c.version)
			assert.NoError(t, err, "must wrap content")

			resp, err := json.Marshal(wrapped)
			assert.NoError(t, err, "must marshal")

			assert.Equal(t, c.expect, string(resp))
		})
	}
}

func TestV2Wrap2(t *testing.T) {

}
