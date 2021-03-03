package metadata

import (
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

// V1APIResponse is the API content.
type V1APIResponse struct {
	Usage UsageV1 `json:"usage"`
	// Content is the actual data. It's currently limited to a table-type response (list of KV maps), but it doesn't need to be.
	Content types.TableContent `json:"content"`
}

// Types of contact information.
type ContactV1 struct {
	PartnersEmail string `json:"partnersEmail"`
}

// Contains contact and usage information.
type UsageV1 struct {
	Contact       ContactV1 `json:"contact"`
	Documentation string    `json:"documentation"`
	Notice        string    `json:"notice"`
}

var DefaultUsageV1 = UsageV1{
	Contact: ContactV1{
		PartnersEmail: partnersEmail,
	},
	Documentation: documentationURL,
	Notice:        defaultNoticeText,
}

// Wraps a list of tabular data with the default usage stanza.
func V1Wrap(table types.TableContent) JSONData {
	return V1APIResponse{
		Usage:   DefaultUsageV1,
		Content: table,
	}
}
