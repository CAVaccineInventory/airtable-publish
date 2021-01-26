module github.com/CAVaccineInventory/airtable-export/monitoring/freshcf

go 1.15

require (
	github.com/CAVaccineInventory/airtable-export/pipeline v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/functions-framework-go v1.2.0
)

replace github.com/CAVaccineInventory/airtable-export/pipeline => ../../pipeline
