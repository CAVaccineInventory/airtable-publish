package providers

import (
	"context"
	"fmt"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/filter"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/honeycombio/beeline-go"
)

func V1(ctx context.Context, tables *airtable.Tables) (airtable.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "endpoints.providers.V1")
	defer span.Send()

	rawTable, err := tables.GetProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Providers table: %w", err)
	}

	filteredTable := filter.ToAllowedKeys(rawTable, []string{
		"Appointments URL",
		"Last Updated",
		"Phase",
		"Provider",
		"Public Notes",
		"Provider network type",
		"Vaccine info URL",
		"Vaccine locations URL",
	})

	return filteredTable, nil
}
