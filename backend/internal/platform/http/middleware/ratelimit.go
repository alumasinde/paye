package middleware

import (
 "net";"net/http";"sync";"time"
)
type bucket struct{count int;reset time.Time}
type RateLimiter struct{mu sync.Mutex;perMinute int;clients map[string]bucket}
func NewRateLimiter(perMinute int)*RateLimiter{return &RateLimiter{perMinute:perMinute,clients:map[string]bucket{}}}
func (l *RateLimiter) Middleware(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
 host,_,err:=net.SplitHostPort(r.RemoteAddr);if err!=nil{host=r.RemoteAddr};now:=time.Now()
 l.mu.Lock();b:=l.clients[host];if now.After(b.reset){b=bucket{reset:now.Add(time.Minute)}}
 b.count++;l.clients[host]=b;allowed:=b.count<=l.perMinute;l.mu.Unlock()
 if !allowed{w.Header().Set("Retry-After","60");http.Error(w,"rate limit exceeded",http.StatusTooManyRequests);return}
 next.ServeHTTP(w,r)
})}
