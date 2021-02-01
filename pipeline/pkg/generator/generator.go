package generator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
	"github.com/honeycombio/beeline-go"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

type StorageWriter func(ctx context.Context, destinationFile string, transformedData airtable.TableContent) error

type PublishManager struct {
	tables *airtable.Tables
	store  StorageWriter
}

func NewPublishManager() *PublishManager {
	return &PublishManager{
		tables: airtable.NewTables(),
		store:  storage.UploadToGCS,
	}
}

func NewDebugPublishManager() *PublishManager {
	return &PublishManager{
		tables: airtable.NewTables(),
		store:  storage.DebugToSTDERR,
	}
}

type TableFetchFunc func(context.Context, string) (airtable.TableContent, error)
type EndpointFunc func(context.Context, *airtable.Tables) (airtable.TableContent, error)
type EndpointMap map[string]EndpointFunc

type unrolledEndpoint struct {
	EndpointName string
	Transform    EndpointFunc
}

func (pm *PublishManager) PublishAll(ctx context.Context, endpoints EndpointMap) bool {
	ctx, span := beeline.StartSpan(ctx, "generator.PublishEndpoint")
	defer span.Send()

	startTime := time.Now()
	wg := sync.WaitGroup{}
	publishOk := make(chan bool, len(endpoints))
	sharedTablesCache := airtable.NewTables()
	for endpointName, transform := range endpoints {
		wg.Add(1)

		endpoint := unrolledEndpoint{
			EndpointName: endpointName,
			Transform:    transform,
		}
		go func(endpoint unrolledEndpoint) {
			defer wg.Done()

			err := pm.PublishEndpoint(ctx, sharedTablesCache, endpoint)
			publishOk <- err == nil
		}(endpoint)
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

func (pm *PublishManager) PublishEndpoint(ctx context.Context, tables *airtable.Tables, endpoint unrolledEndpoint) error {
	ctx, span := beeline.StartSpan(ctx, "generator.PublishEndpoint")
	defer span.Send()
	beeline.AddField(ctx, "endpoint", endpoint.EndpointName)

	tableStartTime := time.Now()
	ctx, _ = tag.New(ctx, tag.Insert(KeyEndpoint, endpoint.EndpointName))

	err := pm.publishEndpointActual(ctx, tables, endpoint)
	if err == nil {
		stats.Record(ctx, TablePublishSuccesses.M(1))
		beeline.AddField(ctx, "success", 1)
		log.Printf("[%s] Successfully published\n", endpoint.EndpointName)
	} else {
		stats.Record(ctx, TablePublishFailures.M(1))
		beeline.AddField(ctx, "failure", 1)
		log.Printf("[%s] Failed to export and publish: %v\n", endpoint.EndpointName, err)
	}
	stats.Record(ctx, TablePublishLatency.M(time.Since(tableStartTime).Seconds()))
	return err
}

func (pm *PublishManager) publishEndpointActual(ctx context.Context, tables *airtable.Tables, endpoint unrolledEndpoint) error {
	log.Printf("[%s] Transforming data...\n", endpoint.EndpointName)
	sanitizedData, err := endpoint.Transform(ctx, tables)
	if err != nil {
		return fmt.Errorf("failed to sanitize json data: %w", err)
	}

	bucket, err := deploys.GetUploadURL(deploys.LegacyVersion)
	if err != nil {
		return fmt.Errorf("failed to get destination bucket: %w", err)
	}
	destinationFile := bucket + "/" + endpoint.EndpointName + ".json"
	log.Printf("[%s] Publishing to %s...\n", endpoint.EndpointName, destinationFile)
	return pm.store(ctx, destinationFile, sanitizedData)
}
