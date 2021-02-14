package counties

import (
	"context"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints/legacy"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	"github.com/honeycombio/beeline-go"
)

func V1(ctx context.Context, tables *airtable.Tables) (types.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "endpoints.counties.V1")
	defer span.Send()

	return legacy.Counties(ctx, tables)
}
