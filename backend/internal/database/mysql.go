package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alumasinde/budget254-paye-api/internal/config"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
)

type MySQL struct{ DB *sql.DB }

func Open(ctx context.Context, c config.DatabaseConfig) (*MySQL, error) {
	q := url.Values{}
	q.Set("parseTime", "true")
	q.Set("charset", "utf8mb4")
	q.Set("loc", "UTC")
	q.Set("timeout", c.ConnectTimeout.String())
	q.Set("readTimeout", "10s")
	q.Set("writeTimeout", "10s")
	if c.TLS {
		q.Set("tls", "true")
	} else {
		q.Set("tls", "false")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.User, c.Password, c.Host, c.Port, c.Name, q.Encode())
	db, e := sql.Open("mysql", dsn)
	if e != nil {
		return nil, fmt.Errorf("open mysql: %w", e)
	}
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	ping, cancel := context.WithTimeout(ctx, c.ConnectTimeout)
	defer cancel()
	if e = db.PingContext(ping); e != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", e)
	}
	return &MySQL{db}, nil
}
func (m *MySQL) Ping(ctx context.Context) error {
	if m == nil || m.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return m.DB.PingContext(ctx)
}
func (m *MySQL) Close() error {
	if m == nil || m.DB == nil {
		return nil
	}
	return m.DB.Close()
}
