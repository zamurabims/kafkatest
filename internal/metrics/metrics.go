package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RecordsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "trending_records_total",
		Help: "Total number of search events processed.",
	})

	RecordsDropped = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "trending_records_dropped_total",
		Help: "Total number of dropped events by reason (stoplist, invalid).",
	}, []string{"reason"})

	TopRequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "trending_top_request_duration_seconds",
		Help:    "Latency of GET /top requests.",
		Buckets: prometheus.DefBuckets,
	})

	CacheRebuildDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "trending_cache_rebuild_duration_seconds",
		Help:    "Time spent rebuilding the top cache.",
		Buckets: prometheus.DefBuckets,
	})

	CacheSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "trending_cache_size",
		Help: "Number of unique queries in the current cached top.",
	})

	StopListSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "trending_stoplist_size",
		Help: "Number of words in the stop list.",
	})
)

func init() {
	prometheus.MustRegister(
		RecordsTotal,
		RecordsDropped,
		TopRequestDuration,
		CacheRebuildDuration,
		CacheSize,
		StopListSize,
	)
}
