package observability

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPMetrics struct {
	registry             *prometheus.Registry
	requests             *prometheus.CounterVec
	duration             *prometheus.HistogramVec
	inflight             prometheus.Gauge
	CatalogSearchResults *prometheus.CounterVec
}

func NewHTTPMetrics(service string) *HTTPMetrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	metrics := &HTTPMetrics{
		registry:             registry,
		requests:             prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cartlabs_http_requests_total", Help: "HTTP requests completed."}, []string{"service", "method", "route", "status"}),
		duration:             prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "cartlabs_http_request_duration_seconds", Help: "HTTP request duration.", Buckets: prometheus.DefBuckets}, []string{"service", "method", "route"}),
		inflight:             prometheus.NewGauge(prometheus.GaugeOpts{Name: "cartlabs_http_requests_in_flight", Help: "HTTP requests currently in flight.", ConstLabels: prometheus.Labels{"service": service}}),
		CatalogSearchResults: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cartlabs_catalog_search_total", Help: "Catalog searches by service or compatibility fallback."}, []string{"result"}),
	}
	registry.MustRegister(metrics.requests, metrics.duration, metrics.inflight, metrics.CatalogSearchResults)
	return metrics
}

func (m *HTTPMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

func (m *HTTPMetrics) Middleware(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		m.inflight.Inc()
		defer m.inflight.Dec()
		next.ServeHTTP(writer, r)
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		} else if _, path, found := strings.Cut(route, " "); found {
			route = path
		}
		m.requests.WithLabelValues(service, r.Method, route, strconv.Itoa(writer.status)).Inc()
		m.duration.WithLabelValues(service, r.Method, route).Observe(time.Since(started).Seconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	written, err := w.ResponseWriter.Write(body)
	w.bytes += written
	return written, err
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type WorkerMetrics struct {
	Registry            *prometheus.Registry
	OutboxRelays        *prometheus.CounterVec
	NotificationResults *prometheus.CounterVec
	SearchResults       *prometheus.CounterVec
	ExpiredReservations prometheus.Counter
}

func NewWorkerMetrics() *WorkerMetrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	metrics := &WorkerMetrics{
		Registry:            registry,
		OutboxRelays:        prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cartlabs_outbox_relay_total", Help: "Outbox relay attempts by result."}, []string{"result"}),
		NotificationResults: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cartlabs_notification_delivery_total", Help: "Notification consumer deliveries by result."}, []string{"result"}),
		SearchResults:       prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cartlabs_search_delivery_total", Help: "Search index consumer deliveries by result."}, []string{"result"}),
		ExpiredReservations: prometheus.NewCounter(prometheus.CounterOpts{Name: "cartlabs_reservations_expired_total", Help: "Expired checkout reservations."}),
	}
	registry.MustRegister(metrics.OutboxRelays, metrics.NotificationResults, metrics.SearchResults, metrics.ExpiredReservations)
	return metrics
}

func (m *WorkerMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

func (m *WorkerMetrics) RegisterOutboxGauges(pool *pgxpool.Pool) {
	m.Registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "cartlabs_outbox_pending", Help: "Outbox events ready or waiting for retry."}, func() float64 {
		return countOutbox(pool, "published_at IS NULL AND dead_lettered_at IS NULL")
	}))
	m.Registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "cartlabs_outbox_dead_letters", Help: "Outbox events requiring explicit replay."}, func() float64 {
		return countOutbox(pool, "dead_lettered_at IS NOT NULL")
	}))
}

func countOutbox(pool *pgxpool.Pool, predicate string) float64 {
	var count int64
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM outbox_events WHERE "+predicate).Scan(&count); err != nil {
		return math.NaN()
	}
	return float64(count)
}
