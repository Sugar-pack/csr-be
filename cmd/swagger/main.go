package main

import (
	"context"
	"os/signal"
	"syscall"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	internalDB "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/db"
	entMigrate "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/migrate"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/logger"
)

func main() {
	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGILL)
	defer cancelFunc()

	lg, err := logger.Get()
	if err != nil {
		lg.Fatal("load config error", zap.Error(err))
	}

	conf, err := config.GetAppConfig()
	if err != nil {
		lg.Fatal("fail to setup app config", zap.Error(err))
	}

	entClient, db, err := internalDB.GetDB(conf.DB.GetConnectionString())
	if err != nil {
		lg.Fatal("failed to db connection", zap.Error(err))
	}

	if conf.DB.EntMigrations {
		if err := entClient.Schema.Create(ctx, entMigrate.WithDropIndex(true)); err != nil {
			lg.Fatal("failed creating schema resources", zap.Error(err))
		}
	}

	if err := internalDB.ApplyMigrations(db); err != nil {
		lg.Fatal("failed to apply migrations", zap.Error(err))
	}

	// setup swagger api
	server, checker, err := SetupAPI(entClient, lg, conf)
	if err != nil {
		lg.Fatal("error setup swagger api", zap.Error(err))
	}

	go checker.PeriodicalCheckup(ctx, conf.OrderStatusOverdueTimeCheckDuration, entClient, lg)

	// Swagger servers handles signals and gracefully shuts down by itself
	if err := server.Serve(); err != nil {
		lg.Error("server fatal error", zap.Error(err))
		return
	}

	if errShutdown := server.Shutdown(); errShutdown != nil {
		lg.Error("error shutting down server", zap.Error(errShutdown))
		return
	}
}
