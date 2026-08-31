package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Database  DatabaseConfig
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Log       LogConfig
}
type AppConfig struct{ Environment, Name, Version string }
type HTTPConfig struct {
	Host                                                                       string
	Port                                                                       int
	ReadTimeout, ReadHeaderTimeout, WriteTimeout, IdleTimeout, ShutdownTimeout time.Duration
	MaxRequestBodyBytes                                                        int64
	TrustProxyHeaders                                                          bool
}
type DatabaseConfig struct {
	Host                                             string
	Port                                             int
	Name, User, Password                             string
	TLS                                              bool
	MaxOpenConns, MaxIdleConns                       int
	ConnMaxLifetime, ConnMaxIdleTime, ConnectTimeout time.Duration
}
type CORSConfig struct {
	AllowedOrigins, AllowedMethods, AllowedHeaders []string
	MaxAgeSeconds                                  int
}
type RateLimitConfig struct {
	Enabled    bool
	Requests   int
	Window     time.Duration
	MaxClients int
}
type LogConfig struct{ Level string }

func Load() (Config, error) {
	var c Config
	var err error
	c.App.Environment = env("APP_ENV", "development")
	c.App.Name = env("APP_NAME", "Budget254 PAYE API")
	c.App.Version = env("APP_VERSION", "0.1.0")
	c.HTTP.Host = env("HTTP_HOST", "0.0.0.0")
	if c.HTTP.Port, err = envInt("HTTP_PORT", 8080); err != nil {
		return c, err
	}
	if c.HTTP.ReadTimeout, err = envDuration("HTTP_READ_TIMEOUT", 10*time.Second); err != nil {
		return c, err
	}
	if c.HTTP.ReadHeaderTimeout, err = envDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second); err != nil {
		return c, err
	}
	if c.HTTP.WriteTimeout, err = envDuration("HTTP_WRITE_TIMEOUT", 15*time.Second); err != nil {
		return c, err
	}
	if c.HTTP.IdleTimeout, err = envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second); err != nil {
		return c, err
	}
	if c.HTTP.ShutdownTimeout, err = envDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second); err != nil {
		return c, err
	}
	if c.HTTP.MaxRequestBodyBytes, err = envInt64("HTTP_MAX_REQUEST_BODY_BYTES", 1048576); err != nil {
		return c, err
	}
	if c.HTTP.TrustProxyHeaders, err = envBool("HTTP_TRUST_PROXY_HEADERS", false); err != nil {
		return c, err
	}
	c.Database.Host = env("DB_HOST", "127.0.0.1")
	if c.Database.Port, err = envInt("DB_PORT", 3306); err != nil {
		return c, err
	}
	c.Database.Name = env("DB_NAME", "budget254_paye")
	c.Database.User = env("DB_USER", "")
	c.Database.Password = os.Getenv("DB_PASSWORD")
	if c.Database.TLS, err = envBool("DB_TLS", false); err != nil {
		return c, err
	}
	if c.Database.MaxOpenConns, err = envInt("DB_MAX_OPEN_CONNS", 20); err != nil {
		return c, err
	}
	if c.Database.MaxIdleConns, err = envInt("DB_MAX_IDLE_CONNS", 10); err != nil {
		return c, err
	}
	if c.Database.ConnMaxLifetime, err = envDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute); err != nil {
		return c, err
	}
	if c.Database.ConnMaxIdleTime, err = envDuration("DB_CONN_MAX_IDLE_TIME", 2*time.Minute); err != nil {
		return c, err
	}
	if c.Database.ConnectTimeout, err = envDuration("DB_CONNECT_TIMEOUT", 5*time.Second); err != nil {
		return c, err
	}
	c.CORS.AllowedOrigins = envList("CORS_ALLOWED_ORIGINS", nil)
	c.CORS.AllowedMethods = envList("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	c.CORS.AllowedHeaders = envList("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-Request-ID"})
	if c.CORS.MaxAgeSeconds, err = envInt("CORS_MAX_AGE_SECONDS", 600); err != nil {
		return c, err
	}
	if c.RateLimit.Enabled, err = envBool("RATE_LIMIT_ENABLED", true); err != nil {
		return c, err
	}
	if c.RateLimit.Requests, err = envInt("RATE_LIMIT_REQUESTS", 60); err != nil {
		return c, err
	}
	if c.RateLimit.Window, err = envDuration("RATE_LIMIT_WINDOW", time.Minute); err != nil {
		return c, err
	}
	if c.RateLimit.MaxClients, err = envInt("RATE_LIMIT_MAX_CLIENTS", 10000); err != nil {
		return c, err
	}
	c.Log.Level = strings.ToUpper(env("LOG_LEVEL", "INFO"))
	return c, c.Validate()
}
func (c Config) Validate() error {
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return errors.New("invalid HTTP_PORT")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return errors.New("invalid DB_PORT")
	}
	if c.Database.Name == "" || c.Database.User == "" {
		return errors.New("DB_NAME and DB_USER are required")
	}
	if c.Database.MaxOpenConns < 1 || c.Database.MaxIdleConns < 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return errors.New("invalid database pool")
	}
	if c.HTTP.MaxRequestBodyBytes < 1 {
		return errors.New("invalid request body limit")
	}
	if c.RateLimit.Enabled && (c.RateLimit.Requests < 1 || c.RateLimit.Window <= 0 || c.RateLimit.MaxClients < 1) {
		return errors.New("invalid rate limit")
	}
	return nil
}
func env(k, f string) string {
	if v, ok := os.LookupEnv(k); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return f
}
func envInt(k string, f int) (int, error) {
	v := env(k, strconv.Itoa(f))
	n, e := strconv.Atoi(v)
	if e != nil {
		return 0, fmt.Errorf("%s: %w", k, e)
	}
	return n, nil
}
func envInt64(k string, f int64) (int64, error) {
	v := env(k, strconv.FormatInt(f, 10))
	n, e := strconv.ParseInt(v, 10, 64)
	if e != nil {
		return 0, fmt.Errorf("%s: %w", k, e)
	}
	return n, nil
}
func envBool(k string, f bool) (bool, error) {
	v := env(k, strconv.FormatBool(f))
	n, e := strconv.ParseBool(v)
	if e != nil {
		return false, fmt.Errorf("%s: %w", k, e)
	}
	return n, nil
}
func envDuration(k string, f time.Duration) (time.Duration, error) {
	v := env(k, f.String())
	n, e := time.ParseDuration(v)
	if e != nil {
		return 0, fmt.Errorf("%s: %w", k, e)
	}
	return n, nil
}
func envList(k string, f []string) []string {
	v, ok := os.LookupEnv(k)
	if !ok || strings.TrimSpace(v) == "" {
		return f
	}
	var o []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			o = append(o, p)
		}
	}
	return o
}
