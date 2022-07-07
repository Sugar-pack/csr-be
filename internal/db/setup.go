package db

import (
	"database/sql"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v4/stdlib"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

func Setup(dbConfig *config.DB) (*sql.DB, *ent.Client, error) {
	connectionString := fmt.Sprintf("host=%s user=csr password=csr dbname=csr sslmode=disable", dbConfig.Host)
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, nil, err
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	entClient := ent.NewClient(ent.Driver(drv))
	return db, entClient, nil
}
