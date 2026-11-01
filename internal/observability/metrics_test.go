package observability

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsExposesRuntimeAndHTTPSeries(t *testing.T) {
	m := NewMetrics("1.0.0-test", "community")
	m.Start()
	m.Finish("/healthz", MethodLabel(http.MethodGet), http.StatusOK, 15*time.Millisecond)
	m.Start()
	m.Finish("unmatched", MethodLabel("CUSTOM_VERB"), http.StatusNotFound, 5*time.Millisecond)

	rr := httptest.NewRecorder()
	m.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rr.Code)
	}
	body := rr.Body.String()
	for _, expected := range []string{
		`moaj_build_info{edition="community",version="1.0.0-test"} 1`,
		`moaj_process_start_time_seconds `,
		`moaj_http_requests_in_flight 0`,
		`moaj_http_requests_total{method="GET",route="/healthz",status="200"} 1`,
		`moaj_http_requests_total{method="OTHER",route="unmatched",status="404"} 1`,
		`moaj_http_request_duration_seconds_count{method="GET",route="/healthz"} 1`,
		`go_goroutines `,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("metrics output missing %q:\n%s", expected, body)
		}
	}
}

func TestNewLoggerOutputsStructuredJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, slog.LevelInfo)
	logger.Debug("debug-suppressed")
	logger.Info("service-started", "edition", "community")
	out := buf.String()
	if strings.Contains(out, "debug-suppressed") {
		t.Fatal("debug message was logged at INFO level")
	}
	if !strings.Contains(out, `"msg":"service-started"`) || !strings.Contains(out, `"edition":"community"`) {
		t.Fatalf("unexpected JSON log output: %s", out)
	}
}
