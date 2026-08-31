package health

import (
 "context";"database/sql";"encoding/json";"net/http";"time"
)
type Handler struct{DB *sql.DB}
func (h Handler) Live(w http.ResponseWriter,r *http.Request){json.NewEncoder(w).Encode(map[string]string{"status":"ok"})}
func (h Handler) Ready(w http.ResponseWriter,r *http.Request){
 ctx,cancel:=context.WithTimeout(r.Context(),2*time.Second);defer cancel()
 if h.DB!=nil {if err:=h.DB.PingContext(ctx);err!=nil{w.WriteHeader(http.StatusServiceUnavailable);json.NewEncoder(w).Encode(map[string]string{"status":"not_ready"});return}}
 json.NewEncoder(w).Encode(map[string]string{"status":"ready"})
}
