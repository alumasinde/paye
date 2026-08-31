package middleware

import (
    "net/http"
    "strings"
)

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
        w.Header().Set("X-Content-Type-Options","nosniff")
        w.Header().Set("X-Frame-Options","DENY")
        w.Header().Set("Referrer-Policy","strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy","camera=(), microphone=(), geolocation=()")
        if r.TLS!=nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"),"https") {
            w.Header().Set("Strict-Transport-Security","max-age=31536000; includeSubDomains")
        }
        next.ServeHTTP(w,r)
    })
}
