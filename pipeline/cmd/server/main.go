package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/deploys"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/generator"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/metrics"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/secrets"
	"github.com/CAVaccineInventory/airtable-export/pipeline/pkg/storage"
	beeline "github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
	"github.com/honeycombio/beeline-go/wrappers/hnynethttp"
	"go.opencensus.io/tag"
)

type Publisher struct {
	lastPublishSucceeded bool // Make this thread safe if nontrivial multithreading comes up.
	deploy               deploys.DeployType
}

// Takes the Google Cloud Storage bucket path as the first argument.
func main() {
	noopFlag := flag.Bool("noop", false, "Only print output, don't upload")
	bucketFlag := flag.String("bucket", "", "Upload into a specific bucket")
	metricsFlag := flag.Bool("metrics", true, "Enable metrics reporting")
	flag.Parse()

	if *noopFlag && *bucketFlag != "" {
		log.Fatal("-noop and -bucket are mutually exclusive!")
	}

	secrets.RequireAirtableSecret()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println()
		log.Printf("Got signal %v, exiting...\n", sig)
		os.Exit(0)
	}()

	if *metricsFlag {
		metricsCleanup := metrics.Init()
		defer metricsCleanup()
	}

	if *noopFlag {
		// "bucket-name" is arbitrary here, since nothing is written anywhere
		deploys.SetTestingStorage(storage.DebugToSTDERR, "bucket-name")
	} else if *bucketFlag != "" {
		deploys.SetTestingStorage(storage.UploadToGCS, *bucketFlag)
	}

	p := Publisher{}
	log.Println("Starting...")
	p.Run()
}

// Run loops forever, publishing data on a regular basis.
func (p *Publisher) Run() {
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

	ctx, _ = tag.New(ctx, tag.Insert(metrics.KeyDeploy, string(p.deploy)))

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
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

	pm := generator.NewPublishManager()
	p.lastPublishSucceeded = pm.PublishAll(ctx)
	if !p.lastPublishSucceeded {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
