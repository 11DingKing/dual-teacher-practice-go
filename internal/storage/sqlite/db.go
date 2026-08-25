package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	_ "modernc.org/sqlite"
)

//go:embed migration.sql
var migrationFS embed.FS

type DB struct{ SQL *sql.DB }

func Open(ctx context.Context, path string) (*DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(8)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	result := &DB{SQL: db}
	if err := result.Migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return result, nil
}
func (d *DB) Close() error { return d.SQL.Close() }
func (d *DB) Migrate(ctx context.Context) error {
	if _, err := d.SQL.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)"); err != nil {
		return err
	}
	var count int
	if err := d.SQL.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version=1").Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		b, err := migrationFS.ReadFile("migration.sql")
		if err != nil {
			return err
		}
		if _, err = d.SQL.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
		if _, err = d.SQL.ExecContext(ctx, "INSERT INTO schema_migrations(version, applied_at) VALUES(1, datetime('now'))"); err != nil {
			return err
		}
	}
	return nil
}
