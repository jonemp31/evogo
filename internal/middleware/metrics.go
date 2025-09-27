package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "evolution_api_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "evolution_api_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Instance metrics
	instancesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "evolution_api_instances_total",
			Help: "Total number of WhatsApp instances",
		},
		[]string{"status"},
	)

	// Message metrics
	messagesSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "evolution_api_messages_sent_total",
			Help: "Total number of messages sent",
		},
		[]string{"instance", "message_type"},
	)

	messagesReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "evolution_api_messages_received_total",
			Help: "Total number of messages received",
		},
		[]string{"instance", "message_type"},
	)

	// Webhook metrics
	webhooksSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "evolution_api_webhooks_sent_total",
			Help: "Total number of webhooks sent",
		},
		[]string{"instance", "event_type", "status"},
	)

	webhookDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "evolution_api_webhook_duration_seconds",
			Help:    "Webhook request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"instance", "event_type"},
	)
)

// MetricsMiddleware collects HTTP metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Extract metrics labels
		method := c.Request.Method
		endpoint := c.FullPath()
		statusCode := strconv.Itoa(c.Writer.Status())

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}

// UpdateInstanceMetrics updates instance-related metrics
func UpdateInstanceMetrics(status string, count int) {
	instancesTotal.WithLabelValues(status).Set(float64(count))
}

// RecordMessageSent records a sent message metric
func RecordMessageSent(instance, messageType string) {
	messagesSentTotal.WithLabelValues(instance, messageType).Inc()
}

// RecordMessageReceived records a received message metric
func RecordMessageReceived(instance, messageType string) {
	messagesReceivedTotal.WithLabelValues(instance, messageType).Inc()
}

// RecordWebhookSent records a webhook sent metric
func RecordWebhookSent(instance, eventType, status string) {
	webhooksSentTotal.WithLabelValues(instance, eventType, status).Inc()
}

// RecordWebhookDuration records webhook duration metric
func RecordWebhookDuration(instance, eventType string, duration float64) {
	webhookDuration.WithLabelValues(instance, eventType).Observe(duration)
}
