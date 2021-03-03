package metadata

import (
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

type Metadata struct {
	Usage UsageV2 `json:"usage"`
}

// Types of contact information.
type ContactV2 struct {
	PartnersEmail string `json:"partners_email"`
}

// Contains contact and usage information.
type UsageV2 struct {
	Contact       ContactV2 `json:"contact"`
	Documentation string    `json:"documentation"`
	Notice        string    `json:"notice"`
}

// V2APIResponse is the API content.
type V2APIResponse struct {
	Metadata Metadata `json:"metadata"`
	// Content is the actual data. It's currently limited to a table-type response (list of KV maps), but it doesn't need to be.
	Content types.TableContent `json:"content"`
}

var DefaultUsageV2 = UsageV2{
	Contact: ContactV2{
		PartnersEmail: partnersEmail,
	},
	Documentation: documentationURL,
	Notice:        defaultNoticeText,
}

// V2Wrap wraps a list of tabular data with the default usage stanza.
func V2Wrap(table types.TableContent) JSONData {
	return V2APIResponse{
		Metadata: Metadata{
			Usage: DefaultUsageV2,
		},
		Content: table,
	}
}
