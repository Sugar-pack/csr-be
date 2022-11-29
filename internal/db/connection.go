package db

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v4/stdlib"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
)

func GetDB(connectionString string) (*ent.Client, *sql.DB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sql connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping sql connection: %w", err)
	}

	entClient := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))

	return entClient, db, nil
}
