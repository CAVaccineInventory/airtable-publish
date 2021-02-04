package airtable

import (
	"context"
	"sync"

	beeline "github.com/honeycombio/beeline-go"
)

// Tables allows just-in-time table fetching and caching from Airtable.
// It is not intended for long-term use, as data is fetched and cached exactly once.
type Tables struct {
	mainLock   sync.RWMutex            // mainLock protects tableLocks.
	tableLocks map[string]*sync.Mutex  // tableLocks contains a lock for each table, to prevent races to populate a table.
	tables     map[string]TableContent // Tables contains a map of table name to table content.
	fetchFunc  func(context.Context, string) (TableContent, error)
}

func NewTables() *Tables {
	return &Tables{
		mainLock:   sync.RWMutex{},
		tableLocks: map[string]*sync.Mutex{},
		tables:     map[string]TableContent{},
		fetchFunc:  Download,
	}
}

// GetTable does a thread-safe, just-in-time fetch of a table.
// The result is cached for the lifetime of the Tables object..
func (t *Tables) GetTable(ctx context.Context, tableName string) (TableContent, error) {
	ctx, span := beeline.StartSpan(ctx, "airtable.GetTable")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)
	// Acquire the lock for the table in question, in order to fetch exactly once or wait for that fetch.
	tableLock := t.getTableLock(tableName)
	tableLock.Lock()
	defer tableLock.Unlock()

	if table, found := t.tables[tableName]; found {
		beeline.AddField(ctx, "fetched", 0)
		return table, nil
	}

	beeline.AddField(ctx, "fetched", 1)
	table, err := t.fetchFunc(ctx, tableName)
	if err != nil {
		beeline.AddField(ctx, "error", err)
		return TableContent{}, err
	}
	t.tables[tableName] = table
	return table, nil
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
