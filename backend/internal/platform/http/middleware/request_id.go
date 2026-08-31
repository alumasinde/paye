package middleware

import (
    "context"
    "net/http"
    "github.com/google/uuid"
)
type contextKey string
const requestIDKey contextKey="request_id"
func RequestID(next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
    id:=r.Header.Get("X-Request-ID");if id==""{id=uuid.NewString()}
    w.Header().Set("X-Request-ID",id)
    next.ServeHTTP(w,r.WithContext(context.WithValue(r.Context(),requestIDKey,id)))
 })
}
func GetRequestID(ctx context.Context)string{v,_:=ctx.Value(requestIDKey).(string);return v}
