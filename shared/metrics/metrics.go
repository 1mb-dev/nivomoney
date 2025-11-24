package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector holds all Prometheus metrics for a service
type Collector struct {
	// HTTP Metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Business Metrics
	TransactionsTotal     *prometheus.CounterVec
	TransactionAmount     *prometheus.HistogramVec
	WalletOperationsTotal *prometheus.CounterVec
	LedgerEntriesTotal    *prometheus.CounterVec
	RiskEventsTotal       *prometheus.CounterVec

	// System Metrics
	DBConnectionsActive prometheus.Gauge
	DBQueryDuration     *prometheus.HistogramVec
	CacheHitsTotal      *prometheus.CounterVec
	CacheMissesTotal    *prometheus.CounterVec
}

// NewCollector creates a new metrics collector for a service
func NewCollector(serviceName string) *Collector {
	return &Collector{
		// HTTP Metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "endpoint", "status"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"service", "method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"service", "method", "endpoint", "status"},
		),

		// Business Metrics
		TransactionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "transactions_total",
				Help: "Total number of transactions",
			},
			[]string{"service", "type", "status"},
		),
		TransactionAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "transaction_amount_inr",
				Help:    "Transaction amount in INR (paise)",
				Buckets: prometheus.ExponentialBuckets(100, 10, 10), // 100 paise to 100M paise
			},
			[]string{"service", "type"},
		),
		WalletOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wallet_operations_total",
				Help: "Total number of wallet operations",
			},
			[]string{"service", "operation", "status"},
		),
		LedgerEntriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ledger_entries_total",
				Help: "Total number of ledger entries",
			},
			[]string{"service", "entry_type", "status"},
		),
		RiskEventsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "risk_events_total",
				Help: "Total number of risk events",
			},
			[]string{"service", "rule", "action"},
		),

		// System Metrics
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"service", "query_type"},
		),
		CacheHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"service", "cache_name"},
		),
		CacheMissesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"service", "cache_name"},
		),
	}
}

// Middleware returns an HTTP middleware that instruments requests
func (c *Collector) Middleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code and size
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Record request size
			if r.ContentLength > 0 {
				c.HTTPRequestSize.WithLabelValues(
					serviceName,
					r.Method,
					r.URL.Path,
				).Observe(float64(r.ContentLength))
			}

			// Process request
			next.ServeHTTP(rw, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(rw.statusCode)

			c.HTTPRequestsTotal.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
				status,
			).Inc()

			c.HTTPRequestDuration.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
				status,
			).Observe(duration)

			c.HTTPResponseSize.WithLabelValues(
				serviceName,
				r.Method,
				r.URL.Path,
				status,
			).Observe(float64(rw.bytesWritten))
		})
	}
}

// Handler returns the Prometheus HTTP handler for /metrics endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Flush implements http.Flusher to support streaming responses like SSE.
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// RecordTransaction records a transaction metric
func (c *Collector) RecordTransaction(serviceName, txType, status string, amountPaise int64) {
	c.TransactionsTotal.WithLabelValues(serviceName, txType, status).Inc()
	c.TransactionAmount.WithLabelValues(serviceName, txType).Observe(float64(amountPaise))
}

// RecordWalletOperation records a wallet operation metric
func (c *Collector) RecordWalletOperation(serviceName, operation, status string) {
	c.WalletOperationsTotal.WithLabelValues(serviceName, operation, status).Inc()
}

// RecordLedgerEntry records a ledger entry metric
func (c *Collector) RecordLedgerEntry(serviceName, entryType, status string) {
	c.LedgerEntriesTotal.WithLabelValues(serviceName, entryType, status).Inc()
}

// RecordRiskEvent records a risk event metric
func (c *Collector) RecordRiskEvent(serviceName, rule, action string) {
	c.RiskEventsTotal.WithLabelValues(serviceName, rule, action).Inc()
}

// RecordDBQuery records a database query duration
func (c *Collector) RecordDBQuery(serviceName, queryType string, duration time.Duration) {
	c.DBQueryDuration.WithLabelValues(serviceName, queryType).Observe(duration.Seconds())
}

// UpdateDBConnections updates the active database connections gauge
func (c *Collector) UpdateDBConnections(count int) {
	c.DBConnectionsActive.Set(float64(count))
}

// RecordCacheHit records a cache hit
func (c *Collector) RecordCacheHit(serviceName, cacheName string) {
	c.CacheHitsTotal.WithLabelValues(serviceName, cacheName).Inc()
}

// RecordCacheMiss records a cache miss
func (c *Collector) RecordCacheMiss(serviceName, cacheName string) {
	c.CacheMissesTotal.WithLabelValues(serviceName, cacheName).Inc()
}
