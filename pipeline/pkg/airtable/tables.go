package airtable

import (
	"context"
	"fmt"
	"sync"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/filter"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/types"
	beeline "github.com/honeycombio/beeline-go"
)

type tableFetchResults struct {
	table types.TableContent
	err   error
}

// Tables allows just-in-time table fetching and caching from Airtable.
// It is not intended for long-term use, as data is fetched and cached exactly once.
type Tables struct {
	mainLock   sync.RWMutex                 // mainLock protects tableLocks.
	tableLocks map[string]*sync.Mutex       // tableLocks contains a lock for each table, to prevent races to populate a table.
	tables     map[string]tableFetchResults // Tables contains a map of table name to (table content or error).
	fetcher    fetcher
}

type fetcher interface {
	Download(context.Context, string) (types.TableContent, error)
}

func NewTables(secret string) *Tables {
	return &Tables{
		mainLock:   sync.RWMutex{},
		tableLocks: map[string]*sync.Mutex{},
		tables:     map[string]tableFetchResults{},
		fetcher:    newAirtable(secret),
	}
}

func (t *Tables) GetCounties(ctx context.Context) (types.TableContent, error) {
	return t.getTable(ctx, "Counties", filter.WithMunger(dropEmpty))
}

func (t *Tables) GetProviders(ctx context.Context) (types.TableContent, error) {
	return t.getTable(ctx, "Provider networks", filter.WithMunger(dropEmpty))
}

func dropEmpty(row map[string]interface{}) (map[string]interface{}, error) {
	// Real rows from airtable will have at least 2 fields.  The synthetic
	// record id field and one real field.
	if len(row) < 2 {
		return nil, nil
	}
	return row, nil
}

func hideNotes(row map[string]interface{}) (map[string]interface{}, error) {
	// Because this function is used as part of the input processing, which only
	// happens once and inside a lock, it directly modifies the input row.
	if v, ok := row["Latest report yes?"].(float64); !ok || v != 1 {
		row["Latest report notes"] = ""
	}
	return row, nil
}

func dropSoftDeleted(row map[string]interface{}) (map[string]interface{}, error) {
	if v, ok := row["is_soft_deleted"].(bool); ok && v {
		return nil, nil
	}
	return row, nil
}

func useCountyURL(ctx context.Context, t *Tables) (func(row map[string]interface{}) (map[string]interface{}, error), error) {
	cs, err := t.GetCounties(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetCounties: %v", err)
	}
	urls := make(map[string]string)
	for _, c := range cs {
		var n, u string
		var ok bool
		if n, ok = c["County"].(string); !ok {
			continue
		}
		if u, ok = c["County vaccination reservations URL"].(string); ok {
			urls[n] = u
		}
	}
	return func(row map[string]interface{}) (map[string]interface{}, error) {
		if county, ok := row["County"].(string); ok {
			if inst, ok := row["Appointment scheduling instructions"].(string); ok {
				if inst == "Uses county scheduling system" {
					if u, ok := urls[county]; ok {
						row["Appointment scheduling instructions"] = u
					}
				}
			}
		}
		return row, nil
	}, nil
}

func (t *Tables) GetLocations(ctx context.Context) (types.TableContent, error) {
	cm, err := useCountyURL(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("Can't setup useCountyURL: %v", err)
	}
	return t.getTable(ctx, "Locations", filter.WithMunger(dropEmpty), filter.WithMunger(hideNotes), filter.WithMunger(dropSoftDeleted), filter.WithMunger(cm))
}

// getTable does a thread-safe, just-in-time fetch of a table.
// The result is cached for the lifetime of the Tables object..
func (t *Tables) getTable(ctx context.Context, tableName string, xfOpts ...filter.XformOpt) (types.TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "airtable.getTable")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)
	// Acquire the lock for the table in question, in order to fetch exactly once or wait for that fetch.
	tableLock := t.getTableLock(tableName)
	tableLock.Lock()
	defer tableLock.Unlock()

	if fetchResult, found := t.tables[tableName]; found {
		beeline.AddField(ctx, "fetched", 0)
		return fetchResult.table, fetchResult.err
	}

	beeline.AddField(ctx, "fetched", 1)
	table, err := t.fetcher.Download(ctx, tableName)
	if err != nil {
		beeline.AddField(ctx, "error", err)
	} else {
		if len(xfOpts) > 0 {
			table, err = filter.Transform(table, xfOpts...)
			if err != nil {
				err = fmt.Errorf("Transform failed: %v", err)
				beeline.AddField(ctx, "error", err)
			}
		}
	}

	t.tables[tableName] = tableFetchResults{
		table: table,
		err:   err,
	}
	return table, err
}

// Returns the lock for the specified table.
// Creates it if it doesn't exist.
func (t *Tables) getTableLock(tableName string) *sync.Mutex {
	t.mainLock.Lock()
	defer t.mainLock.Unlock()

	lock, found := t.tableLocks[tableName]
	if found {
		return lock
	}

	lock = &sync.Mutex{}
	t.tableLocks[tableName] = lock
	return lock
}
