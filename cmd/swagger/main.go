package main

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/db"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/logger"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/migration"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
)

func main() {
	ctx := context.Background()
	logger, err := logger.Setup()
	if err != nil {
		logger.Fatal("load config error", zap.Error(err))
	}

	dbConfig := config.NewDBConfig()
	sqlDB, entClient, err := db.Setup(dbConfig)
	if err != nil {
		logger.Fatal("cant open db", zap.Error(err))
	}

	err = migration.Apply(ctx, sqlDB, entClient)
	if err != migrate.ErrNoChange && err != nil {
		logger.Fatal("migration failed", zap.Error(err))
	}
	logger.Error("migration error", zap.Error(err))

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

	if errServe := server.Serve(); errServe != nil {
		logger.Error("server fatal error", zap.Error(errServe))
		return
	}

	if errShutdown := server.Shutdown(); errShutdown != nil {
		logger.Error("error shutting down server", zap.Error(errShutdown))
		return
	}
}
