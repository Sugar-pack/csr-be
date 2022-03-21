package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/handlers"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
	"github.com/go-openapi/loads"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
)

func main() {
	var loggerConfig = zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(zap.DebugLevel)

	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalln(err)
	}

	dbHost := getEnv("DB_HOST", "localhost")

	connectionString := "host=" + dbHost + " user=csr password=csr dbname=csr sslmode=disable"
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	ctx := context.Background()

	// Run the auto migration tool.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"csr", driver)
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatal(err)
		}
		log.Println(err)
	}
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		logger.Error("error loading swagger spec", zap.Error(err))
		return
	}

	userHandler := handlers.NewUser(
		client,
		logger,
	)

	kindsHandler := handlers.NewKind(
		client,
		logger,
	)
	statusHandler := handlers.NewStatus(
		client,
		logger,
	)

	api := operations.NewBeAPI(swaggerSpec)
	api.UseSwaggerUI()
	api.BearerAuth = middlewares.BearerAuthenticateFunc("key", logger)

	api.UsersPostUserHandler = userHandler.PostUserFunc()
	api.UsersGetCurrentUserHandler = userHandler.GetUserFunc()
	api.UsersPatchUserHandler = userHandler.PatchUserFunc()

	api.KindsCreateNewKindHandler = kindsHandler.CreateNewKindFunc()
	api.KindsGetKindByIDHandler = kindsHandler.GetKindByIDFunc()
	api.KindsDeleteKindHandler = kindsHandler.DeleteKindFunc()
	api.KindsGetAllKindsHandler = kindsHandler.GetAllKindsFunc()

	api.StatusPostStatusHandler = statusHandler.PostStatusFunc()
	api.StatusGetStatusesHandler = statusHandler.GetStatusesFunc()
	api.StatusGetStatusHandler = statusHandler.GetStatusFunc()
	api.StatusDeleteStatusHandler = statusHandler.DeleteStatusFunc()

	server := restapi.NewServer(api)
	listeners := []string{"http"}

	server.EnabledListeners = listeners
	server.Host = getEnv("SERVER_HOST", "127.0.0.1")
	server.Port = 8080

	if err := server.Serve(); err != nil {
		logger.Error("server fatal error", zap.Error(err))
		return
	}

	if err := server.Shutdown(); err != nil {
		logger.Error("error shutting down server", zap.Error(err))
		return
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
