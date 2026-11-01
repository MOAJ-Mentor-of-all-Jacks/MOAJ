package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/enterprise"
	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/observability"
)

type Options struct {
	Logger       *slog.Logger
	Metrics      *observability.Metrics
	MetricsToken string
	MaxBodyBytes int64
	Enterprise   enterprise.Bridge
}

type App struct {
	handler  http.Handler
	draining atomic.Bool
}

func New(options Options) *App {
	a := &App{}
	mux := http.NewServeMux()
	if options.Enterprise != nil {
		options.Enterprise.RegisterRoutes(mux)
	}
	mux.HandleFunc("/healthz", getOnly(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, r, map[string]string{"status": "ok"})
	}))
	mux.HandleFunc("/readyz", getOnly(func(w http.ResponseWriter, r *http.Request) {
		if a.draining.Load() {
			writeProblem(w, r, http.StatusServiceUnavailable, "The service is draining requests.")
			return
		}
		// Phase 1 has no database or queue adapters. Later phases must include
		// readiness checks for every installed dependency.
		writeJSON(w, r, map[string]string{"status": "ready", "scope": "http-runtime"})
	}))
	mux.HandleFunc("/metrics", getOnly(authorizeMetrics(options.MetricsToken, options.Metrics.Handler())))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusNotFound, "No route matches this request.")
	})
	a.handler = middleware(options, mux)
	return a
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) { a.handler.ServeHTTP(w, r) }

func (a *App) BeginShutdown() { a.draining.Store(true) }

func getOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			writeProblem(w, r, http.StatusMethodNotAllowed, "This route accepts GET and HEAD.")
			return
		}
		next.ServeHTTP(w, r)
	}
}

func authorizeMetrics(token string, next http.Handler) http.HandlerFunc {
	expected := sha256.Sum256([]byte(token))
	return func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil || !net.ParseIP(host).IsLoopback() {
				writeProblem(w, r, http.StatusForbidden, "Metrics require a configured access token for remote clients.")
				return
			}
		} else {
			scheme, credentials, found := strings.Cut(r.Header.Get("Authorization"), " ")
			actual := sha256.Sum256([]byte(credentials))
			if !found || !strings.EqualFold(scheme, "Bearer") || subtle.ConstantTimeCompare(expected[:], actual[:]) != 1 {
				w.Header().Set("WWW-Authenticate", `Bearer realm="metrics"`)
				writeProblem(w, r, http.StatusUnauthorized, "A valid metrics access token is required.")
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}

func writeJSON(w http.ResponseWriter, r *http.Request, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_ = json.NewEncoder(w).Encode(value)
	}
}
