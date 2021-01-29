package generator

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"
	"github.com/honeycombio/beeline-go"
)

type FetchManager struct {
	tableData map[string]airtable.Table
}

func (a *FetchManager) FetchAll(ctx context.Context, tableNames []string) {
	ctx, span := beeline.StartSpan(ctx, "generator.FetchAll")
	defer span.Send()

	a.tableData = make(map[string]airtable.Table)

	wg := sync.WaitGroup{}
	for _, tableName := range tableNames {
		wg.Add(1)
		go func(tableName string) {
			defer wg.Done()

			baseTempDir, err := ioutil.TempDir("", tableName)
			defer os.RemoveAll(baseTempDir)
			if err != nil {
				log.Printf("failed to make base temp directory: %v", err)
				return
			}
			jsonMap, err := airtable.Download(ctx, tableName)
			if err != nil {
				log.Printf("failed to fetch from airtable: %v", err)
				return
			}
			a.tableData[tableName] = jsonMap
		}(tableName)
	}

	log.Println("Waiting for all tables to finish fetching...")
	wg.Wait()
	log.Println("All tables finished fetching!")
}

func (a *FetchManager) GetTable(ctx context.Context, tableName string) (airtable.Table, error) {
	ctx, span := beeline.StartSpan(ctx, "generator.FetchTable")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	val, ok := a.tableData[tableName]
	if ok {
		return val, nil
	}
	return nil, errors.New("failed to fetch table")
}
