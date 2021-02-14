package metadata

import (
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

// V1APIResponse is the API content.
type V1APIResponse struct {
	Usage Usage `json:"usage"`
	// Content is the actual data. It's currently limited to a table-type response (list of KV maps), but it doesn't need to be.
	Content types.TableContent `json:"content"`
}

// Wraps a list of tabular data with the default usage stanza.
func V1Wrap(table types.TableContent) JSONData {
	return V1APIResponse{
		Usage:   DefaultUsage,
		Content: table,
	}
}
