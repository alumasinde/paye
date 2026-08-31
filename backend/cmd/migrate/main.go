// Command migrate applies SQL files from ./migrations (default: ../migrations,
// relative to the backend module root) in ascending numeric-prefix order
// (001_, 002_, ...), tracking what has already run in a schema_migrations
// table so re-runs are safe.
//
// The migration files themselves are database-name-agnostic (no CREATE
// DATABASE or USE statements) — this tool owns that: it first connects to
// the MySQL server without selecting a database and creates DB_NAME if
// missing, then reconnects with DB_NAME in the DSN for everything else.
// Set DB_NAME in .env to whatever database you want; it's fully respected.
//
// Usage:
//
//	go run ./cmd/migrate up
//	go run ./cmd/migrate status
//
// Env:
//
//	Same DB_* variables as the API server (see .env.example). A .env file
//	in the working directory is loaded automatically (real environment
//	variables always take priority over it).
//	MIGRATIONS_DIR overrides the default "../migrations" path.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/alumasinde/budget254-paye-api/internal/config"
	"github.com/alumasinde/budget254-paye-api/internal/envfile"
)

var prefixRe = regexp.MustCompile(`^(\d+)_`)

func main() {
	if err := envfile.Load(".env"); err != nil {
		log.Fatalf(".env: %v", err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Deliberately no database name in this first DSN: the target database
	// may not exist yet on a fresh install, and this connection's only job
	// is to create it. A `USE` on a pooled *sql.DB is not guaranteed to
	// carry over to the next Exec (different pooled connections don't share
	// session state), so we don't rely on that - we open a second,
	// database-scoped connection below for everything else.
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&charset=utf8mb4",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port)

	admin, err := sql.Open("mysql", adminDSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := admin.Ping(); err != nil {
		admin.Close()
		log.Fatalf("ping db: %v (check DB_HOST/DB_USER/DB_PASSWORD in .env)", err)
	}

	dbName := cfg.Database.Name
	if _, err := admin.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName,
	)); err != nil {
		admin.Close()
		log.Fatalf("create database %s: %v", dbName, err)
	}
	admin.Close()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&parseTime=true&charset=utf8mb4",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, dbName)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("ping db: %v (check DB_HOST/DB_USER/DB_PASSWORD in .env)", err)
	}

	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id VARCHAR(255) PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT UTC_TIMESTAMP()
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		log.Fatalf("create schema_migrations: %v", err)
	}

	dir := migrationsDir()
	files, err := migrationFiles(dir)
	if err != nil {
		log.Fatalf("read migrations (%s): %v", dir, err)
	}

	applied, err := appliedSet(conn)
	if err != nil {
		log.Fatalf("read applied: %v", err)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "status":
		for _, f := range files {
			mark := "pending"
			if applied[f] {
				mark = "applied"
			}
			fmt.Printf("%-45s %s\n", f, mark)
		}
	case "up":
		ran := 0
		for _, f := range files {
			if applied[f] {
				continue
			}
			if err := applyMigration(conn, dir, f); err != nil {
				log.Fatalf("apply %s: %v", f, err)
			}
			fmt.Printf("applied %s\n", f)
			ran++
		}
		if ran == 0 {
			fmt.Println("nothing to apply; database is up to date")
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q (use: up | status)\n", cmd)
		os.Exit(1)
	}
}

func migrationsDir() string {
	if v := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); v != "" {
		return v
	}
	return "../migrations"
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if !prefixRe.MatchString(e.Name()) {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Slice(files, func(i, j int) bool {
		return prefixNum(files[i]) < prefixNum(files[j])
	})
	return files, nil
}

// prefixNum returns the numeric migration prefix as an int so that, unlike
// a plain string sort, "10_x" correctly sorts after "9_x" regardless of
// zero-padding width.
func prefixNum(name string) int {
	m := prefixRe.FindStringSubmatch(name)
	if len(m) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func appliedSet(conn *sql.DB) (map[string]bool, error) {
	rows, err := conn.Query("SELECT id FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func applyMigration(conn *sql.DB, dir, file string) error {
	sqlBytes, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return err
	}
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(string(sqlBytes)); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.Exec("INSERT INTO schema_migrations (id) VALUES (?)", file); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
