package metadata

import "github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"

// Arbitrary data that may be JSON marshalled.
type JSONData interface{}

// Contact information and usage notice.
type Usage struct {
	Notice  string  `json:"notice"`
	Contact Contact `json:"contact"`
}

// Types of contact information.
type Contact struct {
	PartnersEmail string `json:"partnersEmail"`
}

// APIResponse is the API content.
// TODO: add a metadata struct if and when we start including more metadata-like fields.
type APIResponse struct {
	Usage Usage `json:"usage"`
	// Content is the actual data. It's currently limited to a table-type response (list of KV maps), but it doesn't need to be.
	Content types.TableContent `json:"content"`
}

var defaultNoticeText = "Please contact VaccinateCA and let us know if you plan to rely on or publish this data. This data is provided with best-effort accuracy. If you are displaying this data, we expect you to display it responsibly. Please do not display it in a way that is easy to misread."

// Wraps a list of tabular data with the default usage stanza.
func Wrap(table types.TableContent) JSONData {
	return APIResponse{
		Usage: Usage{
			Notice: defaultNoticeText,
			Contact: Contact{
				PartnersEmail: "api@vaccinateca.com",
			},
		},
		Content: table,
	}
}
