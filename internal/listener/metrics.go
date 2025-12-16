package listener

import "github.com/prometheus/client_golang/prometheus"

// Define metrics
var (
	listenerCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "listener_requests_total",
			Help: "Total number of requests processed by the listener",
		},
		[]string{"status", "func"},
	)
	listenerDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "listener_requests_duration_seconds",
			Help:    "Duration of request processing by the listener",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"func"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(listenerCounter)
	prometheus.MustRegister(listenerDuration)
}
