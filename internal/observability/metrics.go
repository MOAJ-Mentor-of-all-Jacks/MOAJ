package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics owns a registry per application instance, avoiding process-global
// registrations and permitting isolated tests and multiple HTTP servers.
type Metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

func NewMetrics(version, edition string) *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "moaj_http_requests_total", Help: "Completed HTTP requests by bounded route, method and status.",
		}, []string{"route", "method", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "moaj_http_request_duration_seconds", Help: "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "moaj_http_requests_in_flight", Help: "HTTP requests currently being handled.",
		}),
	}
	build := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "moaj_build_info", Help: "Build and edition metadata.",
		ConstLabels: prometheus.Labels{"version": version, "edition": edition},
	})
	build.Set(1)
	startTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "moaj_process_start_time_seconds", Help: "Process start time in seconds since the Unix epoch.",
	})
	startTime.Set(float64(time.Now().Unix()))
	m.registry.MustRegister(m.requests, m.duration, m.inFlight, build, startTime,
		collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{EnableOpenMetrics: true})
}

func (m *Metrics) Start() { m.inFlight.Inc() }

func (m *Metrics) Finish(route, method string, status int, elapsed time.Duration) {
	m.inFlight.Dec()
	m.requests.WithLabelValues(route, method, strconv.Itoa(status)).Inc()
	m.duration.WithLabelValues(route, method).Observe(elapsed.Seconds())
}

// MethodLabel prevents arbitrary HTTP method strings creating unbounded series.
func MethodLabel(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace:
		return method
	default:
		return "OTHER"
	}
}
