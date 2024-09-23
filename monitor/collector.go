package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"metric"},
	)
	CacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"metric"},
	)
	ResponseTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "response_time_seconds",
			Help: "Response time in seconds",
		},
		[]string{"metric"},
	)
)

func Init() {
	prometheus.MustRegister(CacheHits)
	prometheus.MustRegister(CacheMisses)
	prometheus.MustRegister(ResponseTimeHistogram)
}
