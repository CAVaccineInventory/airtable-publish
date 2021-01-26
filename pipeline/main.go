package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const airtableSecretEnvKey = "AIRTABLE_KEY"
const airtableID = "appy2N9zQSnFRPcN8"

var tableNames = [...]string{"Locations", "Counties"}

const tempDir = "airtable-raw"
const readyDir = "airtable-publish"

type Publisher struct {
	bucketPath           string
	lastPublishSucceeded bool // Make this thread safe if nontrivial multithreading comes up.
}

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	log.Println("Starting...")

	p := Publisher{
		bucketPath: os.Args[1],
	}
	p.Run()
}

// Run loops forever, publishing data on a regular basis.
func (p *Publisher) Run() {
	// Serve health status.
	http.HandleFunc("/", p.healthStatus)
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()
	log.Println("Serving health check endpoint...")

	ctx := context.Background()
	// Loop forever.
	for {
		log.Println("Preparing to fetch and publish...")

		// Every iteration gets its own timeout.
		// TODO: re-evaluate 10 minute timeout.
		ctx, cxl := context.WithTimeout(ctx, 10*time.Minute)
		defer cxl()

		// Kick off ETL for each table in parallel.
		// TODO: consider making each table pipeline independent, so a particularly "slow" table export or upload doesn't needlessly delay others.
		wg := sync.WaitGroup{}
		publishOk := make(chan bool, len(tableNames)) // Add a buffer large enough to hold all results.
		for _, tableName := range tableNames {
			wg.Add(1)
			go func(tableName string) {
				defer wg.Done()

				publishErr := p.syncAndPublish(ctx, tableName)
				if publishErr == nil {
					log.Printf("[%s] Successfully published\n", tableName)
				} else {
					log.Printf("[%s] Failed to export and publish: %v\n", tableName, publishErr)
				}
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
		p.lastPublishSucceeded = allPublishOk
		log.Println("All tables finished publishing.")

		time.Sleep(time.Second * 30) // TODO: possibly speed up or make this snazzier.
	}
}

// syncAndPublish fetches data from Airtable, does any necessary transforms/cleanup, then publishes the file to Google Cloud Storage.
// This should probably be broken up further.
func (p *Publisher) syncAndPublish(ctx context.Context, tableName string) error {
	fetchErr := fetchAirtableTable(ctx, tableName)
	if fetchErr != nil {
		return fetchErr
	}

	log.Printf("[%s] Transforming data...\n", tableName)
	j := path.Join(tempDir, tableName+".json")
	jsonMap, err := ObjectFromFile(tableName, j)
	if err != nil {
		return fmt.Errorf("ObjectFromFile(%q): %w", j, err)
	}
	sanitizedData, sanitizeErr := Sanitize(jsonMap, tableName)
	if sanitizeErr != nil {
		return errors.Wrap(sanitizeErr, "failed to sanitize json data")
	}

	destinationFile := p.bucketPath + "/" + tableName + ".json"
	log.Printf("[%s] Getting ready to publish to %s...\n", tableName, destinationFile)
	_ = os.Mkdir(readyDir, 0644)
	f, err := os.Create(path.Join(readyDir, tableName+".json"))
	if err != nil {
		return errors.Wrap(err, "failed to open destination fail")
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.Write(sanitizedData.Bytes())

	if err != nil {
		return errors.Wrap(err, "failed to open write sanitized json")
	}

	return uploadFile(ctx, tableName, destinationFile)
}

// healthStatus returns HTTP 200 if the last publish cycle succeeded,
// and returns HTTP 500 if the last publish cycle failed.
func (p *Publisher) healthStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("Health check called.")
	time.Sleep(time.Second * 70) // TODO: This is a TERRIBLE HACK around how Google Cloud Run sleeps the process when not handling a request.
	lastPublishSucceeded := p.lastPublishSucceeded
	if !lastPublishSucceeded {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "Last run succeeded: %v", lastPublishSucceeded)
}
