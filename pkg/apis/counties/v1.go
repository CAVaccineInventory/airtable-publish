package counties

import (
	"errors"
	"fmt"
	"github.com/CAVaccineInventory/airtable-export/pkg/apis/apimeta"
	"github.com/CAVaccineInventory/airtable-export/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pkg/table"
)

var countiesAllowList = map[string]struct{}{
	"County":                              {},
	"Vaccine info URL":                    {},
	"Vaccine locations URL":               {},
	"Notes":                               {},
	"Total reports":                       {},
	"Yeses":                               {},
	"Official volunteering opportunities": {},
	"Facebook Page":                       {},
}

// CountiesV1 defines the endpoint /counties/v1/counties.
var CountiesV1 = apimeta.EndpointDefinition{
	Group: group,
	Kind:  "counties",
	GenerateResponse: func(tables map[string]table.Table) (apimeta.List, error) {
		countiesTable, found := tables[countiesTableName]
		if !found {
			return apimeta.List{}, errors.New(fmt.Sprintf("%s table not found", countiesTableName))
		}

		filter.FilterToAllowedKeys(countiesTable, countiesAllowList)

		return apimeta.List{
			Metadata: apimeta.Metadata{
				ApiVersion: apimeta.ApiVersion{
					Major:     1,
					Minor:     0,
					Stability: "unstable",
				},
				Contact:     apimeta.DefaultContact,
				UsageNotice: "TODO",
			},
			Content: countiesTable,
		}, nil
	},
}
