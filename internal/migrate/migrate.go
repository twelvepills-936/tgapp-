package migrate

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const advisoryLockKey int64 = 71238461

// Apply runs pending SQL migrations from dir (Flyway-style V{version}__{name}.sql).
func Apply(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if strings.TrimSpace(dir) == "" {
		dir = "internal/migrations"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []migrationFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, desc, ok := parseMigrationName(entry.Name())
		if !ok {
			continue
		}
		files = append(files, migrationFile{
			version: version,
			desc:    desc,
			path:    filepath.Join(dir, entry.Name()),
		})
	}

	if len(files) == 0 {
		return fmt.Errorf("no migrations found in %q", dir)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockKey); err != nil {
		return fmt.Errorf("migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", advisoryLockKey)
	}()

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("schema_migrations table: %w", err)
	}

	for _, file := range files {
		var exists bool
		if err := conn.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`,
			file.version,
		).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		sqlBytes, err := os.ReadFile(file.path)
		if err != nil {
			return err
		}

		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s (%s): %w", file.version, file.desc, err)
		}

		if _, err := conn.Exec(ctx,
			`INSERT INTO schema_migrations (version, description) VALUES ($1, $2)`,
			file.version, file.desc,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", file.version, err)
		}

		slog.InfoContext(ctx, "applied migration",
			slog.String("version", file.version),
			slog.String("description", file.desc),
		)
	}

	return nil
}

type migrationFile struct {
	version string
	desc    string
	path    string
}

func parseMigrationName(name string) (version, description string, ok bool) {
	base := strings.TrimSuffix(name, ".sql")
	if !strings.HasPrefix(base, "V") {
		return "", "", false
	}
	rest := strings.TrimPrefix(base, "V")
	parts := strings.SplitN(rest, "__", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// ResolveDir returns the migrations directory (env MIGRATIONS_DIR or candidates).
func ResolveDir() string {
	if dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); dir != "" {
		return dir
	}
	candidates := []string{
		"internal/migrations",
		"./internal/migrations",
		"/app/internal/migrations",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return "internal/migrations"
}
