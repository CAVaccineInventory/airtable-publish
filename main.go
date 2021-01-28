package main

import (
	"bufio"
	"fmt"
	"github.com/CAVaccineInventory/airtable-export/pkg/apis/legacy"
	"github.com/CAVaccineInventory/airtable-export/pkg/apis/locations"
	"github.com/CAVaccineInventory/airtable-export/pkg/loadjson"
	"github.com/CAVaccineInventory/airtable-export/pkg/sanitize"
	"github.com/CAVaccineInventory/airtable-export/pkg/table"
	"github.com/CAVaccineInventory/airtable-export/pkg/apis/apimeta"
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

// NOTE: Airtable reports a rate limit of 5 calls/second. Depending on how it's validated, we might hit this when adding more tables.
// If this is a concern, slightly stagger the timing of each fetch call.
var tableNames = [...]string{"Locations", "Counties"}

var apiEndpints = []apimeta.EndpointDefinition{
	locations.LocationsV1,
}

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
		p.publishAll()
		time.Sleep(time.Second * 30) // TODO: possibly speed up or make this snazzier.
	}
}

// publishAll publishes updates for all APIs and versions.
func (p *Publisher) publishAll() {
	log.Println("Preparing to fetch and publish...")

	// Populate a map of table name -> table content.
	tables := map[string]table.Table{}
	tablesMutex := sync.Mutex{}

	// Fetch all required tables in parallel.
	wg := sync.WaitGroup{}
	for _, tableName := range tableNames {
		wg.Add(1)
		go func(tableName string) {
			defer wg.Done()
			filePath, err := fetchAirtableTable(tableName)
			if err != nil {
				log.Printf("Failed to fetch table %s: %s\n", tableName, err)
				return
			}
			tableObj, err := loadjson.TableFromJson(filePath)
			if err != nil {
				log.Printf("Failed to marshal table %s: %s\n", tableName, err)
				return
			}

			// Safely write the table.
			tablesMutex.Lock()
			tables[tableName] = tableObj
			tablesMutex.Unlock()
		}(tableName)
	}

	wg.Wait()
	log.Println("All required tables fetched.")

	// Generate all API views.
	// TODO: add a waitgroup to make publishAll wait until all api endpoints complete or time out.

	// TODO: sanitize and upload.
	legacyLocations := legacy.Locations(tables["Locations"])
	legacyCounties := legacy.Counties(tables["Counties"])

	wg = sync.WaitGroup{}
	for _, endpoint := range apiEndpints {
		wg.Add(1)

		// TODO: break this up
		go func(endpointDef apimeta.EndpointDefinition) {
			defer wg.Done()

			unsanitizedResponse, err := endpointDef.GenerateResponse(tables)
			if err != nil {
				log.Printf("Failed to generate: %v", err)
			}

			sanitizedResponse, err := sanitize.Sanitize(unsanitizedResponse)
			if err != nil {
				// TODO: handle
			}

			// TODO: write json
			// TODO: upload
		}(endpoint)
	}

	wg.Wait()
	log.Println("Finished publising data.")
}

// syncAndPublish fetches data from Airtable, does any necessary transforms/cleanup, then publishes the file to Google Cloud Storage.
// This should probably be broken up further.
func (p *Publisher) syncAndPublish(tableName string) error {
	rawFilePath, fetchErr := fetchAirtableTable(tableName)
	if fetchErr != nil {
		return fetchErr
	}

	log.Println("Transforming data...")
	jsonMap, err := loadjson.TableFromJson(path.Join(rawFilePath))
	sanitizedData, sanitizeErr := sanitize.Sanitize(jsonMap)
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
