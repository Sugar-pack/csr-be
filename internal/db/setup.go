package db

import (
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

func Setup(logger *zap.Logger) (*sql.DB, *ent.Client) {
	db, err := sql.Open("sqlite3", "file:csr?mode=memory&cache=shared&_fk=1")
	if err != nil {
		logger.Fatal("cant open db", zap.Error(err))
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.SQLite, db)
	entClient := ent.NewClient(ent.Driver(drv))

	return db, entClient
}
