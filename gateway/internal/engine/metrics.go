package engine

import (
	"bufio"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zeromicro/go-zero/core/logx"
)

type metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	latency  *prometheus.HistogramVec
	active   prometheus.Gauge
	reloads  prometheus.Counter
}

func newMetrics() *metrics {
	m := &metrics{registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "gateway_requests_total", Help: "Completed gateway requests."}, []string{"route", "status"}),
		latency:  prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "gateway_request_duration_seconds", Help: "Time including streaming responses.", Buckets: []float64{.005, .025, .1, .5, 1, 5, 30, 120, 600}}, []string{"route"}),
		active:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "gateway_active_requests", Help: "Active proxy requests including upgraded streams."}),
		reloads:  prometheus.NewCounter(prometheus.CounterOpts{Name: "gateway_config_applied_total", Help: "Successful complete snapshot activations."}),
	}
	m.registry.MustRegister(m.requests, m.latency, m.active, m.reloads)
	return m
}
func (e *Engine) Metrics() http.Handler {
	return promhttp.HandlerFor(e.metrics.registry, promhttp.HandlerOpts{})
}

type accessWriter struct {
	http.ResponseWriter
	status, bytes int
}

func (w *accessWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *accessWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, buffered, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err == nil {
		w.status = http.StatusSwitchingProtocols
	}
	return conn, buffered, err
}
func (w *accessWriter) WriteHeader(status int) {
	if status >= 200 || status == 101 {
		if w.status != 0 {
			return
		}
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *accessWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(200)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}
func (e *Engine) observe(w *accessWriter, route string, started time.Time) {
	status := w.status
	if status == 0 {
		status = 200
	}
	e.metrics.active.Dec()
	e.metrics.requests.WithLabelValues(route, strconv.Itoa(status)).Inc()
	e.metrics.latency.WithLabelValues(route).Observe(time.Since(started).Seconds())
	// Route IDs, not raw paths/query strings or credential headers, bound label
	// cardinality and keep credentials out of access logs.
	logx.Infow("gateway request", logx.Field("route", route), logx.Field("status", status), logx.Field("bytes", w.bytes), logx.Field("duration_ms", time.Since(started).Milliseconds()))
}
