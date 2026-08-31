package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/alumasinde/budget254-paye-api/internal/config"
	"github.com/alumasinde/budget254-paye-api/internal/response"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type key string

const requestIDKey key = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if !validID(id) {
			id = randomID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}
func RequestIDFromContext(c context.Context) string { v, _ := c.Value(requestIDKey).(string); return v }
func validID(v string) bool {
	if len(v) < 8 || len(v) > 128 {
		return false
	}
	for _, r := range v {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
func randomID() string {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
func Recovery(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if x := recover(); x != nil {
					l.Error("panic recovered", "request_id", RequestIDFromContext(r.Context()), "panic", x)
					response.Error(w, 500, "INTERNAL_ERROR", "an unexpected error occurred", RequestIDFromContext(r.Context()))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
func MaxBodyBytes(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}
func CORS(c config.CORSConfig) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range c.AllowedOrigins {
		allowed[o] = true
	}
	methods := strings.Join(c.AllowedMethods, ", ")
	headers := strings.Join(c.AllowedHeaders, ", ")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			o := r.Header.Get("Origin")
			if o != "" && allowed[o] {
				w.Header().Set("Access-Control-Allow-Origin", o)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", methods)
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}
			if r.Method == http.MethodOptions {
				if o != "" && !allowed[o] {
					response.Error(w, 403, "ORIGIN_NOT_ALLOWED", "origin not allowed", RequestIDFromContext(r.Context()))
					return
				}
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type window struct {
	start time.Time
	count int
}

func RateLimit(c config.RateLimitConfig, trust bool) func(http.Handler) http.Handler {
	var mu sync.Mutex
	clients := map[string]window{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !c.Enabled {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r, trust)
			now := time.Now()
			mu.Lock()
			v := clients[ip]
			if v.start.IsZero() || now.Sub(v.start) >= c.Window {
				v = window{start: now}
			}
			if v.count >= c.Requests {
				mu.Unlock()
				response.Error(w, 429, "RATE_LIMITED", "too many requests, please try again shortly", RequestIDFromContext(r.Context()))
				return
			}
			v.count++
			clients[ip] = v
			if len(clients) > c.MaxClients {
				for k, x := range clients {
					if now.Sub(x.start) >= c.Window {
						delete(clients, k)
					}
				}
			}
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
func clientIP(r *http.Request, trust bool) string {
	if trust {
		if x := r.Header.Get("X-Forwarded-For"); x != "" {
			f := strings.TrimSpace(strings.Split(x, ",")[0])
			if net.ParseIP(f) != nil {
				return f
			}
		}
	}
	h, _, e := net.SplitHostPort(r.RemoteAddr)
	if e == nil {
		return h
	}
	return r.RemoteAddr
}
