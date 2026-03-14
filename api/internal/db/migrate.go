package db

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate reads all SQL files from the embedded migrations directory,
// tracks which have been applied in a schema_migrations table, and
// runs any pending ones in filename order.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// Ensure tracking table exists
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Read all migration files
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Sort by filename (001_, 002_, etc.)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		name := entry.Name()

		// Check if already applied
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		// Read and execute migration
		sql, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		log.Printf("Applying migration: %s", name)

		_, err = pool.Exec(ctx, string(sql))
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		// Record it
		_, err = pool.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`, name)
		if err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		log.Printf("Applied migration: %s", name)
	}

	return nil
}
