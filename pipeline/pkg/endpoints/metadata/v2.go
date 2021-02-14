package metadata

import (
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

type Metadata struct {
	Usage Usage `json:"usage"`
}

// V2APIResponse is the API content.
type V2APIResponse struct {
	Metadata Metadata `json:"metadata"`
	// Content is the actual data. It's currently limited to a table-type response (list of KV maps), but it doesn't need to be.
	Content types.TableContent `json:"content"`
}

// V2Wrap wraps a list of tabular data with the default usage stanza.
func V2Wrap(table types.TableContent) JSONData {
	return V2APIResponse{
		Metadata: Metadata{
			Usage: DefaultUsage,
		},
		Content: table,
	}
}
