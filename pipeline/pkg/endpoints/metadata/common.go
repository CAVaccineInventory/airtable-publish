package metadata

import (
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
)

// Versions are, practically, numbers, but we use strings for
// potential extensibility.
type VersionType string

// "Legacy" versions are ones which predate the CDN domain and bucket.
const LegacyVersion VersionType = "LEGACY"

// Arbitrary data that may be JSON marshalled.
type JSONData interface{}

// Contact information and usage notice.
type Usage struct {
	Contact       Contact `json:"contact"`
	Documentation string  `json:"documentation"`
	Notice        string  `json:"notice"`
}

// Types of contact information.
type Contact struct {
	PartnersEmail string `json:"partnersEmail"`
}

const partnersEmail = "api@vaccinateca.com"
const defaultNoticeText = "Please contact VaccinateCA and let us know if you plan to rely on or publish this data. This data is provided with best-effort accuracy. If you are displaying this data, we expect you to display it responsibly. Please do not display it in a way that is easy to misread."
const documentationURL = "https://docs.vaccinateca.com"

var DefaultUsage = Usage{
	Contact: Contact{
		PartnersEmail: partnersEmail,
	},
	Documentation: documentationURL,
	Notice:        defaultNoticeText,
}

// Wrap adds any applicable metadata and API structure, beyond the table content.
func Wrap(tableData types.TableContent, version VersionType) (JSONData, error) {
	switch version {
	case LegacyVersion:
		return tableData, nil // Don't wrap the legacy version, it would break format compatibility with any consumers.
	case "1":
		return V1Wrap(tableData), nil
	case "2":
		return V2Wrap(tableData), nil
	default:
		return nil, fmt.Errorf("unrecognized version %s", version)
	}
}
