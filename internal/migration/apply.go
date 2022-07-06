package migration

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	entMigrate "git.epam.com/epm-lstr/epm-lstr-lc/be/ent/migrate"
)

func Apply(ctx context.Context, db *sql.DB, entClient *ent.Client) error {
	// Run the auto migration tool.
	if err := entClient.Schema.Create(ctx, entMigrate.WithDropIndex(true)); err != nil {
		return err
		//logger.Fatal("failed creating schema resources", zap.Error(err))
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"csr", driver)
	if err != nil {
		return err
	}
	if errUp := m.Up(); errUp != nil {
		return errUp
	}
	return nil
}
