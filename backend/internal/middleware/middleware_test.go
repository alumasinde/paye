package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alumasinde/budget254-paye-api/internal/config"
)

func TestRequestID(t *testing.T) {
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r.Context()) == "" {
			t.Fatal("missing id")
		}
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing response id")
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	h := CORS(config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:8081"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAgeSeconds:  600,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("OPTIONS should not reach the next handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/calculator/paye", nil)
	req.Header.Set("Origin", "http://localhost:8081")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8081" {
		t.Fatalf("allow origin = %q", got)
	}
}

func TestCORSAllowsPrivateNetworkOriginWhenEnabled(t *testing.T) {
	h := CORS(config.CORSConfig{
		AllowPrivateNetworkOrigins: true,
		AllowedMethods:             []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:             []string{"Content-Type"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("OPTIONS should not reach the next handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/calculator/paye", nil)
	req.Header.Set("Origin", "http://192.168.100.11:8081")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://192.168.100.11:8081" {
		t.Fatalf("allow origin = %q", got)
	}
}

func TestCORSRejectsPrivateNetworkOriginWhenDisabled(t *testing.T) {
	h := CORS(config.CORSConfig{
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/calculator/paye", nil)
	req.Header.Set("Origin", "http://192.168.100.11:8081")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}
