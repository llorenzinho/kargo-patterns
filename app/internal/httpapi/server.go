// Package httpapi exposes the service HTTP endpoints.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

const serviceName = "kargo-pattern"

type server struct {
	logger  *slog.Logger
	version string
	commit  string
}

// New returns the HTTP handler for the service.
func New(logger *slog.Logger, version, commit string) http.Handler {
	s := &server{
		logger:  logger,
		version: version,
		commit:  commit,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleRoot)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /readyz", s.handleReady)
	mux.HandleFunc("GET /version", s.handleVersion)

	return s.logRequests(mux)
}

func (s *server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, map[string]string{
		"message": "hello from " + serviceName,
		"service": serviceName,
		"version": s.version,
	})
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, map[string]string{"status": "ok"})
}

func (s *server) handleReady(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, map[string]string{"status": "ok"})
}

func (s *server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, map[string]string{
		"service": serviceName,
		"version": s.version,
		"commit":  s.commit,
	})
}

func (s *server) writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		s.logger.Error("encode response", "error", err)
	}
}

func (s *server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &responseRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}

		s.logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.Int("status", status),
			slog.Int("bytes", recorder.bytes),
			slog.Float64("duration_ms", float64(time.Since(started).Microseconds())/1000),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(p)
	r.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
