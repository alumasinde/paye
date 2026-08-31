package response

import "net/http"

// FailError is the body written by Fail. Named separately from the Error
// function in json.go to avoid a package-level redeclaration.
type FailError struct { Code string `json:"code"`; Message string `json:"message"`; Fields map[string]string `json:"fields,omitempty"`; RequestID string `json:"request_id,omitempty"` }
func Fail(w http.ResponseWriter,status int,code,msg,id string,fields map[string]string){ JSON(w,status,FailError{code,msg,fields,id}) }
