package airtable

import "go.opencensus.io/stats"

var FetchLatency = stats.Float64(
	"airtable_fetch_latency_s",
	"Latency for the airtable-extract phase",
	stats.UnitSeconds,
)
