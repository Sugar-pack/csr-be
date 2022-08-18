package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	entMigrate "git.epam.com/epm-lstr/epm-lstr-lc/be/ent/migrate"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/logger"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	logger, err := logger.Setup()
	if err != nil {
		logger.Fatal("load config error", zap.Error(err))
	}

	go periodicalStopCheck(time.Millisecond*500, logger)

	defer func() {
		_ = recover()
		if err := clearPID(); err != nil {
			logger.Fatal("failed to clear pid file", zap.Error(err))
		}
	}()

	db, err := sql.Open("sqlite3", "file:csr?mode=memory&cache=shared&_fk=1")
	if err != nil {
		logger.Fatal("cant open db", zap.Error(err))
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.SQLite, db)
	entClient := ent.NewClient(ent.Driver(drv))

	// Run the auto migration tool.
	if err := entClient.Schema.Create(ctx, entMigrate.WithDropIndex(true)); err != nil {
		logger.Fatal("failed creating schema resources", zap.Error(err))
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		logger.Fatal("failed to create sqlite3 driver", zap.Error(err))
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"csr", driver)
	if err != nil {
		logger.Fatal("failed to create migrate", zap.Error(err))
	}
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			logger.Fatal("migration failed", zap.Error(err))
		}
		logger.Error("migration error", zap.Error(err))
	}

	// conf
	serverConfig, err := config.SetupServerConfig()
	if err != nil {
		logger.Fatal("fail to setup server config", zap.Error(err))
	}

	appConfig, err := config.SetupAppConfig()
	if err != nil {
		logger.Fatal("fail to setup app config", zap.Error(err))
	}
	// setup swagger api
	api, err := swagger.SetupAPI(entClient, logger, appConfig)
	if err != nil {
		logger.Fatal("error setup swagger api", zap.Error(err))
	}

	// run server
	server := restapi.NewServer(api)
	listeners := []string{"http"}

	server.EnabledListeners = listeners
	server.Host = serverConfig.Host
	server.Port = serverConfig.Port

	if err := writePID(); err != nil {
		logger.Fatal("failed to write pid file", zap.Error(err))
	}

	if err := server.Serve(); err != nil {
		logger.Error("server fatal error", zap.Error(err))
		return
	}

	if errShutdown := server.Shutdown(); errShutdown != nil {
		logger.Error("error shutting down server", zap.Error(errShutdown))
		return
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

const pidFileName = "pid"

func writePID() error {
	return os.WriteFile(pidFileName, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func clearPID() error {
	return os.Remove(pidFileName)
}

func periodicalStopCheck(duration time.Duration, logger *zap.Logger) {
	ticker := time.NewTicker(duration)
	for {
		<-ticker.C

		_, err := os.Stat("stop")
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				logger.Info("failed to check stop file existence, unexpected error", zap.Error(err))
			}
			continue
		}
		// file exists
		logger.Info("stop file exists, exiting")
		// TODO graceful shutdown
		os.Exit(0)
	}
}
