package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/observability"
)

func testOptions(out io.Writer) Options {
	return Options{
		Logger:       observability.NewLogger(out, slog.LevelDebug),
		Metrics:      observability.NewMetrics("test", "community"),
		MetricsToken: "a-long-test-token-never-record-in-logs", MaxBodyBytes: 1024,
	}
}

func TestEndpointsAndReadiness(t *testing.T) {
	app := New(testOptions(io.Discard))
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/healthz", 200},
		{http.MethodHead, "/healthz", 200},
		{http.MethodGet, "/readyz", 200},
		{http.MethodPost, "/healthz", 405},
		{http.MethodGet, "/unknown", 404},
		{http.MethodGet, "/api/v1/tasks", 404},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			app.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.path, nil))
			if rr.Code != tc.status {
				t.Fatalf("status = %d; want %d", rr.Code, tc.status)
			}
			if len(rr.Header().Get("X-Request-ID")) != 32 || rr.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Fatal("missing request ID or security headers")
			}
			if tc.method == http.MethodHead && rr.Body.Len() != 0 {
				t.Fatal("HEAD response must not contain a body")
			}
			if tc.status >= 400 {
				assertProblem(t, rr, tc.status)
			}
			if tc.path == "/readyz" && !strings.Contains(rr.Body.String(), `"scope":"http-runtime"`) {
				t.Fatal("readiness must disclose its current limited scope")
			}
		})
	}
	app.BeginShutdown()
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	assertProblem(t, rr, http.StatusServiceUnavailable)
	rr = httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatal("draining must not mark the live process as dead")
	}
}

func TestMetricsAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name, token, auth, remote, forwarded string
		status                               int
	}{
		{"valid token", "secret", "Bearer secret", "192.0.2.1:5000", "", 200},
		{"missing token", "secret", "", "127.0.0.1:5000", "", 401},
		{"wrong token", "secret", "Bearer invalid", "127.0.0.1:5000", "", 401},
		{"wrong scheme", "secret", "Basic secret", "127.0.0.1:5000", "", 401},
		{"local development", "", "", "127.0.0.1:5000", "", 200},
		{"local ipv6", "", "", "[::1]:5000", "", 200},
		{"remote development", "", "", "192.0.2.1:5000", "", 403},
		{"spoofed proxy header", "", "", "192.0.2.1:5000", "127.0.0.1", 403},
		{"invalid peer", "", "", "invalid", "", 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := testOptions(io.Discard)
			opts.MetricsToken = tc.token
			r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			r.RemoteAddr = tc.remote
			r.Header.Set("Authorization", tc.auth)
			r.Header.Set("X-Forwarded-For", tc.forwarded)
			rr := httptest.NewRecorder()
			New(opts).ServeHTTP(rr, r)
			if rr.Code != tc.status {
				t.Fatalf("got %d, want %d", rr.Code, tc.status)
			}
			if tc.status != 200 {
				assertProblem(t, rr, tc.status)
			} else if !strings.Contains(rr.Body.String(), "go_goroutines") ||
				!strings.Contains(rr.Body.String(), "moaj_build_info") ||
				!strings.Contains(rr.Body.String(), "moaj_process_start_time_seconds") {
				t.Fatal("authenticated scrape did not expose runtime metrics")
			}
		})
	}
}

func TestTelemetryOmitsSecretsAndBoundsLabels(t *testing.T) {
	var logs bytes.Buffer
	opts := testOptions(&logs)
	app := New(opts)
	r := httptest.NewRequest("PRIVATE_METHOD", "/private-personnel?email=secret@example.test", strings.NewReader("private note"))
	r.Header.Set("Authorization", "Bearer "+opts.MetricsToken)
	r.Header.Set("X-Request-ID", "private-client-id")
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, r)
	if rr.Header().Get("X-Request-ID") == "private-client-id" {
		t.Fatal("untrusted request ID was accepted")
	}
	for _, sensitive := range []string{"private-personnel", "secret@example.test", "private note", "private-client-id", opts.MetricsToken, "PRIVATE_METHOD"} {
		if strings.Contains(logs.String(), sensitive) {
			t.Fatalf("access log leaked %q", sensitive)
		}
	}
	scrape := httptest.NewRecorder()
	opts.Metrics.Handler().ServeHTTP(scrape, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(scrape.Body.String(), `moaj_http_requests_total{method="OTHER",route="unmatched",status="404"} 1`) {
		t.Fatalf("unexpected request counter: %s", scrape.Body.String())
	}
	if strings.Contains(scrape.Body.String(), "private-personnel") {
		t.Fatal("metric label contains request path")
	}
	var logRecord map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logs.Bytes()), &logRecord); err != nil {
		t.Fatalf("invalid JSON log: %v", err)
	}
	if logRecord["request_id"] != rr.Header().Get("X-Request-ID") {
		t.Fatal("request and log correlation IDs differ")
	}
}

func TestMatchedRouteMetric(t *testing.T) {
	opts := testOptions(io.Discard)
	New(opts).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
	rr := httptest.NewRecorder()
	opts.Metrics.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(rr.Body.String(), `route="/healthz"`) {
		t.Fatal("matched route template was not recorded")
	}
}

func TestRecoveryDoesNotExposePanic(t *testing.T) {
	var logs bytes.Buffer
	h := middleware(testOptions(&logs), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("database password should never appear")
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/panic", nil))
	assertProblem(t, rr, http.StatusInternalServerError)
	if strings.Contains(logs.String()+rr.Body.String(), "database password") {
		t.Fatal("panic contents were exposed")
	}
}

func TestRecoveryAbortsPartiallyWrittenResponses(t *testing.T) {
	h := middleware(testOptions(io.Discard), http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("partial"))
		panic("private details")
	}))
	defer func() {
		if recover() != http.ErrAbortHandler {
			t.Fatal("partially written response must be aborted")
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/panic", nil))
}

func TestBodyLimits(t *testing.T) {
	for _, unknownLength := range []bool{false, true} {
		opts := testOptions(io.Discard)
		opts.MaxBodyBytes = 3
		called := false
		h := middleware(opts, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			_, err := io.ReadAll(r.Body)
			var limitErr *http.MaxBytesError
			if !errors.As(err, &limitErr) {
				t.Errorf("body reader did not enforce maximum: %v", err)
			}
			writeProblem(w, r, http.StatusRequestEntityTooLarge, "Body too large.")
		}))
		r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("1234"))
		if unknownLength {
			r.ContentLength = -1
		}
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, r)
		assertProblem(t, rr, http.StatusRequestEntityTooLarge)
		if called != unknownLength {
			t.Fatal("known oversized body should be rejected before handler invocation")
		}
	}
}

func assertProblem(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	var p problem
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("invalid problem JSON: %v (%s)", err, rr.Body.String())
	}
	if rr.Code != status || p.Status != status || p.Type != "about:blank" || p.Title == "" || p.RequestID == "" || rr.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("invalid problem response: status=%d body=%+v headers=%v", rr.Code, p, rr.Header())
	}
}
