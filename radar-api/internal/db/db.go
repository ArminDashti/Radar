package db

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ArminDashti/radar-api/internal/auth"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const (
	seedUsername   = "armin"
	seedPassword   = "dopadopa123"
	seedAgentToken = "radar-demo-agent-token-probe1"
	seedMinutes    = 120
)

// Connect opens the SQLite file at path. ponytail: a single connection serializes
// all access (no lock contention); upgrade to a WAL read pool when grid load grows.
func Connect(ctx context.Context, path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	conn.SetMaxOpenConns(1)
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return conn, nil
}

// RunMigration applies the SQL file once, recording it in schema_migrations.
func RunMigration(ctx context.Context, conn *sql.DB, path string) error {
	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	name := filepath.Base(path)
	var applied int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&applied); err != nil {
		return fmt.Errorf("check migration: %w", err)
	}
	if applied > 0 {
		return nil
	}
	sqlText, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, string(sqlText)); err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (name, applied_at) VALUES (?, ?)`, name, time.Now().UTC()); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}
	return tx.Commit()
}

func Seed(ctx context.Context, conn *sql.DB) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO users (username, password_hash, role) VALUES (?, ?, 'admin')
		ON CONFLICT (username) DO UPDATE SET role = 'admin'`, seedUsername, string(passwordHash)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO agents (name, probe_id, token_hash)
		SELECT 'Demo agent probe1', id, ? FROM probes WHERE code = 'probe1'
		ON CONFLICT (name) DO UPDATE SET token_hash = excluded.token_hash`,
		auth.HashAgentToken(seedAgentToken)); err != nil {
		return err
	}
	var agentID int64
	if err = tx.QueryRowContext(ctx, `SELECT id FROM agents WHERE name = 'Demo agent probe1'`).Scan(&agentID); err != nil {
		return err
	}
	if err = seedSamples(ctx, tx, agentID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE agents SET last_seen_at = ? WHERE id = ?`, time.Now().UTC(), agentID); err != nil {
		return err
	}
	return tx.Commit()
}

type seedTarget struct {
	id   int64
	http bool
	icmp bool
}

// seedSamples generates demo history in Go: random() inside a SQLite subquery
// is not guaranteed to be evaluated once per row, which can break the samples CHECK.
func seedSamples(ctx context.Context, tx *sql.Tx, agentID int64) error {
	rows, err := tx.QueryContext(ctx, `SELECT id, http_enabled, icmp_enabled FROM endpoints`)
	if err != nil {
		return err
	}
	var targets []seedTarget
	for rows.Next() {
		var target seedTarget
		if err := rows.Scan(&target.id, &target.http, &target.icmp); err != nil {
			rows.Close()
			return err
		}
		targets = append(targets, target)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	insert, err := tx.PrepareContext(ctx, `
		INSERT INTO samples (agent_id, endpoint_id, protocol, observed_at, latency_ms, ok)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (agent_id, endpoint_id, protocol, observed_at) DO NOTHING`)
	if err != nil {
		return err
	}
	defer insert.Close()

	newest := time.Now().UTC().Truncate(time.Minute)
	for _, target := range targets {
		var protocols []string
		if target.http {
			protocols = append(protocols, "http")
		}
		if target.icmp {
			protocols = append(protocols, "icmp")
		}
		for _, protocol := range protocols {
			for minute := 0; minute < seedMinutes; minute++ {
				observedAt := newest.Add(-time.Duration(seedMinutes-1-minute) * time.Minute).Unix()
				ok := rand.Float64() >= 0.08
				var latency *float64
				if ok {
					value := 20 + rand.Float64()*180
					latency = &value
				}
				if _, err := insert.ExecContext(ctx, agentID, target.id, protocol, observedAt, latency, ok); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// IsUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
// ponytail: string match on the driver message; upgrade to the sqlite error code if needed.
func IsUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// InClause returns "?,?,..." placeholders and matching args for an IN (...) list.
func InClause[T any](values []T) (string, []any) {
	marks := make([]string, len(values))
	args := make([]any, len(values))
	for index, value := range values {
		marks[index] = "?"
		args[index] = value
	}
	return strings.Join(marks, ","), args
}
