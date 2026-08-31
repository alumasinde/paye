package server

import (
 "context";"net/http";"time"
)
func Run(ctx context.Context,addr string,handler http.Handler,shutdown time.Duration)error{
 srv:=&http.Server{Addr:addr,Handler:handler,ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:30*time.Second,IdleTimeout:60*time.Second}
 errc:=make(chan error,1);go func(){errc<-srv.ListenAndServe()}()
 select{case err:=<-errc:return err;case <-ctx.Done():
  c,cancel:=context.WithTimeout(context.Background(),shutdown);defer cancel()
  _=srv.Shutdown(c);return ctx.Err()
 }
}
