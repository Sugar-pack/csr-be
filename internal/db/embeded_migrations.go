package db

import (
	"database/sql"
	"embed"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

func ApplyMigrations(db *sql.DB) error {
	mgs := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: embeddedMigrations,
		Root:       "migrations",
	}
	if _, err := migrate.Exec(db, "postgres", mgs, migrate.Up); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}
