package main

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

const airtableSecretEnvKey = "AIRTABLE_KEY"
const airtableId = "appy2N9zQSnFRPcN8"

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

	// Loop forever.
	for {
		log.Println("Preparing to fetch and publish...")

		// Kick off ETL for each table in parallel.
		// TODO: consider making each table pipeline independent, so a particularly "slow" table export or upload doesn't needlessly delay others.
		wg := sync.WaitGroup{}
		publishOk := make(chan bool, len(tableNames)) // Add a buffer large enough to hold all results.
		for _, tableName := range tableNames {
			wg.Add(1)
			go func(tableName string) {
				defer wg.Done()

				publishErr := p.syncAndPublish(tableName)
				if publishErr == nil {
					log.Printf("Successfully published table %s\n", tableName)
				} else {
					log.Printf("Failed to export and publish table %s: %v\n", tableName, publishErr)
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
func (p *Publisher) syncAndPublish(tableName string) error {
	fetchErr := fetchAirtableTable(tableName)
	if fetchErr != nil {
		return fetchErr
	}

	log.Println("Transforming data...")
	jsonMap, err := ObjectFromFile(path.Join(tempDir, tableName+".json"))
	sanitizedData, sanitizeErr := Sanitize(jsonMap, tableName)
	if sanitizeErr != nil {
		return errors.Wrap(sanitizeErr, "failed to sanitize json data")
	}

	destinationFile := p.bucketPath + "/" + tableName + ".json"
	log.Printf("Getting ready to publish to %s...\n", destinationFile)
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

	return uploadFile(tableName, destinationFile)
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
	fmt.Fprintf(w, "Last run suceeded: %v", lastPublishSucceeded)
}
