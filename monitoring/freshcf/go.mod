module github.com/CAVaccineInventory/airtable-export/monitoring/freshcf

go 1.15

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	github.com/CAVaccineInventory/airtable-export/pipeline v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/functions-framework-go v1.2.0
	go.opencensus.io v0.22.5
)

replace github.com/CAVaccineInventory/airtable-export/pipeline => ../../pipeline
