package middleware

import (
 "log/slog";"net/http";"runtime/debug"
)
func Recover(log *slog.Logger,next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){defer func(){if x:=recover();x!=nil{log.Error("panic recovered","error",x,"stack",string(debug.Stack()));w.WriteHeader(http.StatusInternalServerError);_,_=w.Write([]byte(`{"code":"INTERNAL_ERROR","message":"Internal server error"}`))}}();next.ServeHTTP(w,r)})}
