package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/deploys"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

const airtableSecretEnvKey = "AIRTABLE_KEY"
const airtableID = "appy2N9zQSnFRPcN8"

var tableNames = [...]string{"Locations", "Counties"}

type Publisher struct {
	lastPublishSucceeded bool // Make this thread safe if nontrivial multithreading comes up.
}

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	log.Println("Starting...")

	p := Publisher{}
	p.Run()
}

// Run loops forever, publishing data on a regular basis.
func (p *Publisher) Run() {
	metricsCleanup := InitMetrics()
	defer metricsCleanup()

	// Serve health status.
	http.HandleFunc("/", p.healthStatus)
	http.HandleFunc("/publish", p.syncAndPublishRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func (p *Publisher) syncAndPublishRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	deploy, err := deploys.GetDeploy()
	if err != nil {
		panic(err)
	}
	ctx, _ = tag.New(ctx, tag.Insert(keyDeploy, string(deploy)))
	startTime := time.Now()
	log.Println("Preparing to fetch and publish...")

	// Every iteration gets its own timeout.  Update the README.md
	// for new latencies if you adjust this.
	timeoutMinutes, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil {
		timeoutMinutes = 2
	}
	ctx, cxl := context.WithTimeout(ctx, time.Duration(timeoutMinutes)*time.Minute)
	defer cxl()

	// Kick off ETL for each table in parallel.
	// TODO: consider making each table pipeline independent, so a particularly "slow" table export or upload doesn't needlessly delay others.
	wg := sync.WaitGroup{}
	publishOk := make(chan bool, len(tableNames)) // Add a buffer large enough to hold all results.
	for _, tableName := range tableNames {
		wg.Add(1)
		go func(tableName string) {
			defer wg.Done()

			tableStartTime := time.Now()
			tableCtx, _ := tag.New(ctx, tag.Insert(keyTable, tableName))

			publishErr := p.syncAndPublish(tableCtx, tableName)
			if publishErr == nil {
				stats.Record(tableCtx, tablePublishSuccesses.M(1))
				log.Printf("[%s] Successfully published\n", tableName)
			} else {
				stats.Record(tableCtx, tablePublishFailures.M(1))
				log.Printf("[%s] Failed to export and publish: %v\n", tableName, publishErr)
			}
			stats.Record(tableCtx, tablePublishLatency.M(time.Since(tableStartTime).Seconds()))
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
	stats.Record(ctx, totalPublishLatency.M(time.Since(startTime).Seconds()))
	p.lastPublishSucceeded = allPublishOk
	log.Println("All tables finished publishing.")
	if !allPublishOk {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// syncAndPublish fetches data from Airtable, does any necessary transforms/cleanup, then publishes the file to Google Cloud Storage.
// This should probably be broken up further.
func (p *Publisher) syncAndPublish(ctx context.Context, tableName string) error {
	baseTempDir, err := ioutil.TempDir("", tableName)
	defer os.RemoveAll(baseTempDir)
	if err != nil {
		return fmt.Errorf("failed to make base temp directory: %w", err)
	}
	inDir := path.Join(baseTempDir, "in")
	err = os.Mkdir(inDir, 0644)
	if err != nil {
		return fmt.Errorf("failed to make in directory %s: %w", inDir, err)
	}

	start := time.Now()
	filePath, err := fetchAirtableTable(ctx, inDir, tableName)
	stats.Record(ctx, airtableFetchLatency.M(time.Since(start).Seconds()))

	if err != nil {
		return fmt.Errorf("failed to fetch from airtable: %w", err)
	}

	log.Printf("[%s] Transforming data...\n", tableName)
	jsonMap, err := ObjectFromFile(tableName, filePath)
	if err != nil {
		return fmt.Errorf("failed to parse json in %s: %w", filePath, err)
	}
	sanitizedData, err := Sanitize(jsonMap, tableName)
	if err != nil {
		return fmt.Errorf("failed to sanitize json data: %w", err)
	}

	bucket, err := deploys.GetExportBucket()
	if err != nil {
		return fmt.Errorf("failed to get destination bucket: %w", err)
	}
	localFile := path.Join(baseTempDir, tableName+".json")
	destinationFile := bucket + "/" + tableName + ".json"
	log.Printf("[%s] Getting ready to publish to %s...\n", tableName, destinationFile)
	f, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localFile, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.Write(sanitizedData.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write sanitized json to %s: %w", localFile, err)
	}

	return uploadFile(ctx, localFile, destinationFile)
}

// healthStatus returns HTTP 200 if the last publish cycle succeeded,
// and returns HTTP 500 if the last publish cycle failed.
func (p *Publisher) healthStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("Health check called.")
	lastPublishSucceeded := p.lastPublishSucceeded
	if !lastPublishSucceeded {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "Last run succeeded: %v", lastPublishSucceeded)
}
