package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEndpoints(t *testing.T) {
	t.Parallel()

	handler := New(slog.New(slog.NewTextHandler(io.Discard, nil)), "1.2.3", "abc1234")

	tests := []struct {
		name string
		path string
		want map[string]string
	}{
		{
			name: "root",
			path: "/",
			want: map[string]string{
				"service": serviceName,
				"version": "1.2.3",
				"message": "hello from " + serviceName,
			},
		},
		{
			name: "health",
			path: "/healthz",
			want: map[string]string{"status": "ok"},
		},
		{
			name: "ready",
			path: "/readyz",
			want: map[string]string{"status": "ok"},
		},
		{
			name: "version",
			path: "/version",
			want: map[string]string{
				"service": serviceName,
				"version": "1.2.3",
				"commit":  "abc1234",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Fatalf("content type = %q", got)
			}

			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			for key, want := range tt.want {
				if got := body[key]; got != want {
					t.Fatalf("%s = %q, want %q", key, got, want)
				}
			}
		})
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	handler := New(slog.New(slog.NewTextHandler(io.Discard, nil)), "1.2.3", "abc1234")
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestWrongMethodReturnsMethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := New(slog.New(slog.NewTextHandler(io.Discard, nil)), "1.2.3", "abc1234")
	req := httptest.NewRequest(http.MethodPost, "/version", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
