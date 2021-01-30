package generator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
	"github.com/honeycombio/beeline-go"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

type PublishManager struct {
	FetchManager
}

func NewPublishManager() *PublishManager {
	return &PublishManager{}
}

func (pm *PublishManager) PublishAll(ctx context.Context, tableNames []string) bool {
	ctx, span := beeline.StartSpan(ctx, "generator.Publish")
	defer span.Send()

	pm.FetchAll(ctx, tableNames)

	startTime := time.Now()
	wg := sync.WaitGroup{}
	publishOk := make(chan bool, len(tableNames)) // Add a buffer large enough to hold all results.
	for _, tableName := range tableNames {
		wg.Add(1)
		go func(tableName string) {
			defer wg.Done()

			publishErr := pm.Publish(ctx, tableName)
			publishOk <- publishErr == nil
		}(tableName)
	}

	log.Println("Waiting for all tables to finish publishing...")
	wg.Wait()
	allPublishOk := true
	for len(publishOk) != 0 {
		if !<-publishOk {
			allPublishOk = false
			break
		}
	}
	stats.Record(ctx, TotalPublishLatency.M(time.Since(startTime).Seconds()))
	log.Println("All tables finished publishing.")
	return allPublishOk
}

func (pm *PublishManager) Publish(ctx context.Context, tableName string) error {
	ctx, span := beeline.StartSpan(ctx, "generator.Publish")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	tableStartTime := time.Now()
	ctx, _ = tag.New(ctx, tag.Insert(KeyTable, tableName))

	err := pm.publishActual(ctx, tableName)
	if err == nil {
		stats.Record(ctx, TablePublishSuccesses.M(1))
		beeline.AddField(ctx, "success", 1)
		log.Printf("[%s] Successfully published\n", tableName)
	} else {
		stats.Record(ctx, TablePublishFailures.M(1))
		beeline.AddField(ctx, "failure", 1)
		log.Printf("[%s] Failed to export and publish: %v\n", tableName, err)
	}
	stats.Record(ctx, TablePublishLatency.M(time.Since(tableStartTime).Seconds()))
	return err
}

func (pm *PublishManager) publishActual(ctx context.Context, tableName string) error {
	jsonMap, err := pm.GetTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("failed to fetch json data: %w", err)
	}

	log.Printf("[%s] Transforming data...\n", tableName)
	sanitizedData, err := Transform(ctx, jsonMap, tableName)
	if err != nil {
		return fmt.Errorf("failed to sanitize json data: %w", err)
	}

	return storage.UploadToGCS(ctx, tableName, sanitizedData)
}
