package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	beeline "github.com/honeycombio/beeline-go"
)

func DebugToSTDERR(ctx context.Context, destinationFile string, transformedData airtable.TableContent) error {
	ctx, span := beeline.StartSpan(ctx, "storage.Print")
	defer span.Send()
	beeline.AddField(ctx, "destinationFile", destinationFile)

	serializedData, err := transformedData.Serialize()
	if err != nil {
		return fmt.Errorf("failed to write serialized json: %w", err)
	}

	byteLen := len(serializedData.Bytes())
	fmt.Fprintf(os.Stderr, "======> Would write %d bytes to %s\n", byteLen, destinationFile)
	return err
}
