package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
)

type Config struct {
    Environment string
    HTTPAddress string
    DatabaseDSN string
    CORSOrigins []string
    JWTIssuer string
    JWTAccessSecret string
    RequestTimeout time.Duration
    ShutdownTimeout time.Duration
    RateLimitPerMinute int
    TrustProxy bool
}

func Load() (Config,error) {
    c:=Config{
        Environment:getenv("APP_ENV","development"),
        HTTPAddress:getenv("HTTP_ADDR",":8080"),
        DatabaseDSN:os.Getenv("MYSQL_DSN"),
        CORSOrigins:split(getenv("CORS_ORIGINS","")),
        JWTIssuer:getenv("JWT_ISSUER","budget254-paye"),
        JWTAccessSecret:os.Getenv("JWT_ACCESS_SECRET"),
        RequestTimeout:duration("REQUEST_TIMEOUT","15s"),
        ShutdownTimeout:duration("SHUTDOWN_TIMEOUT","20s"),
        RateLimitPerMinute:intenv("RATE_LIMIT_PER_MINUTE",120),
        TrustProxy:getenv("TRUST_PROXY","false")=="true",
    }
    if c.Environment=="production" {
        if c.DatabaseDSN=="" { return c,fmt.Errorf("MYSQL_DSN is required") }
        if len(c.JWTAccessSecret)<32 { return c,fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 bytes") }
        if len(c.CORSOrigins)==0 { return c,fmt.Errorf("CORS_ORIGINS is required in production") }
    }
    return c,nil
}
func getenv(k,d string)string{if v:=strings.TrimSpace(os.Getenv(k));v!=""{return v};return d}
func split(s string)[]string{var out []string;for _,v:=range strings.Split(s,","){if v=strings.TrimSpace(v);v!=""{out=append(out,v)}};return out}
func duration(k,d string)time.Duration{v:=getenv(k,d);x,err:=time.ParseDuration(v);if err!=nil{return mustDuration(d)};return x}
func mustDuration(v string)time.Duration{x,_:=time.ParseDuration(v);return x}
func intenv(k string,d int)int{v:=getenv(k,"");if v==""{return d};n,err:=strconv.Atoi(v);if err!=nil||n<1{return d};return n}
