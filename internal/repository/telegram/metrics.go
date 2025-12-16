package telegram

import "github.com/prometheus/client_golang/prometheus"

// Define metrics
var (
	messageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telegram_messages_total",
			Help: "Total number of messages processed by the Telegram client",
		},
		[]string{"status", "func"},
	)
	messageDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "telegram_message_duration_seconds",
			Help:    "Duration of message processing by the Telegram client",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"func"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(messageCounter)
	prometheus.MustRegister(messageDuration)
}
