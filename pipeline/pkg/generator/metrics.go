package generator

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	TotalPublishLatency = stats.Float64(
		"total_publish_latency_s",
		"Latency to extract and publish all measures",
		stats.UnitSeconds,
	)

	TablePublishLatency = stats.Float64(
		"table_publish_latency_s",
		"Latency to extract and publish each table",
		stats.UnitSeconds,
	)

	TablePublishSuccesses = stats.Int64(
		"table_publish_successes_count",
		"Number of successful publishes by table",
		stats.UnitDimensionless,
	)

	TablePublishFailures = stats.Int64(
		"table_publish_failures_count",
		"Number of failed publishes by table",
		stats.UnitDimensionless,
	)

	KeyTable, _ = tag.NewKey("table")
)
