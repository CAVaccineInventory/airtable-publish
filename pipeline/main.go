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

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	beeline "github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
	"github.com/honeycombio/beeline-go/wrappers/hnynethttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var tableNames = []string{"Locations", "Counties"}

type Publisher struct {
	lastPublishSucceeded bool // Make this thread safe if nontrivial multithreading comes up.
	tableManager         generator.FetchManager
	deploy               deploys.DeployType
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

	deploy, err := deploys.GetDeploy()
	if err != nil {
		panic(err)
	}
	p.deploy = deploy

	// Serve health status.
	http.HandleFunc("/", p.healthStatus)
	http.HandleFunc("/publish", p.syncAndPublishRequest)
	err = http.ListenAndServe(":8080", hnynethttp.WrapHandler(http.DefaultServeMux))
	if err != nil {
		panic(err)
	}
}

func (p *Publisher) syncAndPublishRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.GetSpanFromContext(ctx)
	span.AddField("handler.name", "syncAndPublishRequest")
	beeline.AddFieldToTrace(ctx, "deploy", string(p.deploy))

	ctx, _ = tag.New(ctx, tag.Insert(keyDeploy, string(p.deploy)))
	startTime := time.Now()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	log.Println("Preparing to fetch and publish...")

	// Every iteration gets its own timeout.  Update the README.md
	// for new latencies if you adjust this.
	timeoutMinutes, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil {
		timeoutMinutes = 2
	}
	ctx, cxl := context.WithTimeout(ctx, time.Duration(timeoutMinutes)*time.Minute)
	defer cxl()

	p.tableManager = generator.FetchManager{}
	p.tableManager.FetchAll(ctx, tableNames)

	wg := sync.WaitGroup{}
	publishOk := make(chan bool, len(tableNames)) // Add a buffer large enough to hold all results.
	for _, tableName := range tableNames {
		wg.Add(1)
		go func(tableName string) {
			defer wg.Done()

			publishErr := p.publish(ctx, tableName)
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

func (p *Publisher) publish(ctx context.Context, tableName string) error {
	ctx, span := beeline.StartSpan(ctx, "publish")
	defer span.Send()
	beeline.AddField(ctx, "table", tableName)

	tableStartTime := time.Now()
	ctx, _ = tag.New(ctx, tag.Insert(keyTable, tableName))

	err := p.publishActual(ctx, tableName)
	if err == nil {
		stats.Record(ctx, tablePublishSuccesses.M(1))
		beeline.AddField(ctx, "success", 1)
		log.Printf("[%s] Successfully published\n", tableName)
	} else {
		stats.Record(ctx, tablePublishFailures.M(1))
		beeline.AddField(ctx, "failure", 1)
		log.Printf("[%s] Failed to export and publish: %v\n", tableName, err)
	}
	stats.Record(ctx, tablePublishLatency.M(time.Since(tableStartTime).Seconds()))
	return err
}

// syncAndPublish fetches data from Airtable, does any necessary transforms/cleanup, then publishes the file to Google Cloud Storage.
// This should probably be broken up further.
func (p *Publisher) publishActual(ctx context.Context, tableName string) error {
	jsonMap, err := p.tableManager.GetTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("failed to fetch json data: %w", err)
	}

	log.Printf("[%s] Transforming data...\n", tableName)
	sanitizedData, err := Sanitize(ctx, jsonMap, tableName)
	if err != nil {
		return fmt.Errorf("failed to sanitize json data: %w", err)
	}

	tempDir, err := ioutil.TempDir("", tableName)
	defer os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("failed to make temp directory: %w", err)
	}
	localFile := path.Join(tempDir, tableName+".json")
	log.Printf("[%s] Getting ready to publish...\n", tableName)
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

	return storage.UploadToGCS(ctx, tableName, localFile)
}

// healthStatus returns HTTP 200 if the last publish cycle succeeded,
// and returns HTTP 500 if the last publish cycle failed.
func (p *Publisher) healthStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.GetSpanFromContext(ctx)
	span.AddField("handler.name", "healthcheck")
	beeline.AddFieldToTrace(ctx, "deploy", string(p.deploy))

	log.Println("Health check called.")
	lastPublishSucceeded := p.lastPublishSucceeded
	if !lastPublishSucceeded {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "Last run succeeded: %v", lastPublishSucceeded)
}
