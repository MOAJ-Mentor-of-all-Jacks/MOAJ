package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/internal/observability"
)

type requestIDKey struct{}

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// middleware logs bounded route/method labels and server-generated IDs only.
// Bodies, queries, headers, addresses and panic values are deliberately omitted.
func middleware(options Options, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var idBytes [16]byte
		_, _ = rand.Read(idBytes[:]) // crypto/rand.Read fills the buffer or terminates.
		id := hex.EncodeToString(idBytes[:])
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id))
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}
		tracked := &responseWriter{ResponseWriter: w}
		started := time.Now()
		options.Metrics.Start()
		defer func() {
			status := tracked.status
			if status == 0 {
				status = http.StatusOK
			}
			route := r.Pattern
			if route == "" || route == "/" {
				route = "unmatched"
			}
			method := observability.MethodLabel(r.Method)
			elapsed := time.Since(started)
			options.Metrics.Finish(route, method, status, elapsed)
			options.Logger.Info("HTTP request", "request_id", id, "method", method,
				"route", route, "status", status, "duration_ms", elapsed.Milliseconds(), "bytes", tracked.bytes)
		}()
		defer func() {
			if recovered := recover(); recovered != nil {
				if err, ok := recovered.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(http.ErrAbortHandler)
				}
				options.Logger.Error("HTTP handler panicked", "request_id", id)
				if tracked.status != 0 {
					// A partial response cannot become a problem document. Let
					// net/http abort it without logging sensitive panic contents.
					panic(http.ErrAbortHandler)
				}
				writeProblem(tracked, r, http.StatusInternalServerError, "The request could not be completed.")
			}
		}()
		if r.ContentLength > options.MaxBodyBytes {
			writeProblem(tracked, r, http.StatusRequestEntityTooLarge, "The request body exceeds the configured limit.")
			return
		}
		r.Body = http.MaxBytesReader(tracked, r.Body, options.MaxBodyBytes)
		next.ServeHTTP(tracked, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseWriter) WriteHeader(status int) {
	if status >= 100 && status < 200 {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytes += n
	return n, err
}

// Unwrap preserves ResponseController support, including future SSE deadlines.
func (w *responseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
