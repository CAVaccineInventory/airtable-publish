package generator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/endpoints"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
	"github.com/honeycombio/beeline-go"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

type StorageWriter func(ctx context.Context, destinationFile string, transformedData airtable.TableContent) error

type PublishManager struct {
	store StorageWriter
}

func NewPublishManager() *PublishManager {
	return &PublishManager{
		store: storage.UploadToGCS,
	}
}

func NewDebugPublishManager() *PublishManager {
	return &PublishManager{
		store: storage.DebugToSTDERR,
	}
}

func (pm *PublishManager) PublishAll(ctx context.Context) bool {
	ctx, span := beeline.StartSpan(ctx, "generator.PublishAll")
	defer span.Send()

	eps := endpoints.AllEndpoints()

	startTime := time.Now()
	wg := sync.WaitGroup{}
	publishOk := make(chan bool, len(eps))
	sharedTablesCache := airtable.NewTables()
	for _, ep := range eps {
		wg.Add(1)

		go func(ep endpoints.Endpoint) {
			defer wg.Done()

			err := pm.PublishEndpoint(ctx, sharedTablesCache, ep)
			publishOk <- err == nil
		}(ep)
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

func (pm *PublishManager) PublishEndpoint(ctx context.Context, tables *airtable.Tables, ep endpoints.Endpoint) error {
	ctx, span := beeline.StartSpan(ctx, "generator.PublishEndpoint")
	defer span.Send()
	beeline.AddField(ctx, "version", ep.Version)
	beeline.AddField(ctx, "resource", ep.Resource)

	tableStartTime := time.Now()
	ctx, _ = tag.New(ctx,
		tag.Insert(KeyVersion, string(ep.Version)),
		tag.Insert(KeyResource, ep.Resource))

	err := pm.publishEndpointActual(ctx, tables, ep)
	if err == nil {
		stats.Record(ctx, TablePublishSuccesses.M(1))
		beeline.AddField(ctx, "success", 1)
		log.Printf("[%v] Successfully published\n", &ep)
	} else {
		stats.Record(ctx, TablePublishFailures.M(1))
		beeline.AddField(ctx, "failure", 1)
		log.Printf("[%v] Failed to export and publish: %v\n", &ep, err)
	}
	stats.Record(ctx, TablePublishLatency.M(time.Since(tableStartTime).Seconds()))
	return err
}

func (pm *PublishManager) publishEndpointActual(ctx context.Context, tables *airtable.Tables, ep endpoints.Endpoint) error {
	log.Printf("[%v] Transforming data...\n", &ep)
	sanitizedData, err := ep.Transform(ctx, tables)
	if err != nil {
		return fmt.Errorf("failed to sanitize json data: %w", err)
	}

	bucket, err := deploys.GetUploadURL(ep.Version)
	if err != nil {
		return fmt.Errorf("failed to get destination bucket: %w", err)
	}
	destinationFile := bucket + "/" + ep.Resource + ".json"
	log.Printf("[%v] Publishing to %s...\n", &ep, destinationFile)
	return pm.store(ctx, destinationFile, sanitizedData)
}
