package middleware
import ("context";"net/http";"strings";"github.com/golang-jwt/jwt/v5";"github.com/alumasinde/budget254-paye-api/internal/response")
type userKey string
const authenticatedUser userKey="auth_user"
func UserID(ctx context.Context)string{v,_:=ctx.Value(authenticatedUser).(string);return v}
func unauthorized(w http.ResponseWriter, r *http.Request){response.Fail(w,401,"UNAUTHORIZED","unauthorized",RequestIDFromContext(r.Context()),nil)}
func RequireAuth(secret []byte,next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){parts:=strings.Fields(r.Header.Get("Authorization"));if len(parts)!=2||parts[0]!="Bearer"{unauthorized(w,r);return};token,err:=jwt.Parse(parts[1],func(t *jwt.Token)(any,error){return secret,nil});if err!=nil||!token.Valid{unauthorized(w,r);return};claims,ok:=token.Claims.(jwt.MapClaims);if !ok||claims["typ"]!="customer"{unauthorized(w,r);return};sub,_:=claims["sub"].(string);if sub==""{unauthorized(w,r);return};next.ServeHTTP(w,r.WithContext(context.WithValue(r.Context(),authenticatedUser,sub)))})}
