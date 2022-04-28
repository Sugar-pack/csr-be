package main

import (
	"context"
	"database/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
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
	os.Setenv("JWT_SECRET_KEY", "123")
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

	equipmentHandler := handlers.NewEquipment(
		client,
		logger,
	)

	userHandler := handlers.NewUser(
		client,
		logger,
	)

	roleHandler := handlers.NewRole(
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
	activeAreasHandler := handlers.NewActiveArea(
		client,
		logger,
	)
	ordersHandler := handlers.NewOrder(
		client,
		logger,
	)

	orderStatus := handlers.NewOrderStatus(
		client,
		logger,
	)

	api := operations.NewBeAPI(swaggerSpec)
	api.UseSwaggerUI()
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Fatalln("JWT_SECRET_KEY not specified")
	}
	api.BearerAuth = middlewares.BearerAuthenticateFunc(jwtSecretKey, logger)

	api.UsersLoginHandler = userHandler.LoginUserFunc(jwtSecretKey)
	api.UsersPostUserHandler = userHandler.PostUserFunc()
	api.UsersGetCurrentUserHandler = userHandler.GetUserFunc()
	api.UsersPatchUserHandler = userHandler.PatchUserFunc()
	api.UsersAssignRoleToUserHandler = userHandler.AssignRoleToUserFunc(repositories.NewUserRepository(client))

	api.RolesGetRolesHandler = roleHandler.GetRolesFunc()

	api.KindsCreateNewKindHandler = kindsHandler.CreateNewKindFunc()
	api.KindsGetKindByIDHandler = kindsHandler.GetKindByIDFunc()
	api.KindsDeleteKindHandler = kindsHandler.DeleteKindFunc()
	api.KindsGetAllKindsHandler = kindsHandler.GetAllKindsFunc()
	api.KindsPatchKindHandler = kindsHandler.PatchKindFunc()

	api.StatusPostStatusHandler = statusHandler.PostStatusFunc()
	api.StatusGetStatusesHandler = statusHandler.GetStatusesFunc()
	api.StatusGetStatusHandler = statusHandler.GetStatusFunc()
	api.StatusDeleteStatusHandler = statusHandler.DeleteStatusFunc()

	api.EquipmentCreateNewEquipmentHandler = equipmentHandler.PostEquipmentFunc()
	api.EquipmentGetEquipmentHandler = equipmentHandler.GetEquipmentFunc()
	api.EquipmentDeleteEquipmentHandler = equipmentHandler.DeleteEquipmentFunc()
	api.EquipmentGetAllEquipmentHandler = equipmentHandler.ListEquipmentFunc()
	api.EquipmentEditEquipmentHandler = equipmentHandler.EditEquipmentFunc()
	api.EquipmentFindEquipmentHandler = equipmentHandler.FindEquipmentFunc()

	api.ActiveAreasGetAllActiveAreasHandler = activeAreasHandler.GetActiveAreasFunc()

	orderRepository := repositories.NewOrderRepository(client)
	api.OrdersGetAllOrdersHandler = ordersHandler.ListOrderFunc(orderRepository)
	api.OrdersCreateOrderHandler = ordersHandler.CreateOrderFunc(orderRepository)
	api.OrdersUpdateOrderHandler = ordersHandler.UpdateOrderFunc(orderRepository)

	orderStatusRepertory := repositories.NewOrderFilter(client)
	api.OrdersGetOrdersByStatusHandler = orderStatus.GetOrdersByStatus(orderStatusRepertory)
	api.OrdersGetOrdersByDateAndStatusHandler = orderStatus.GetOrdersByPeriodAndStatus(orderStatusRepertory)

	statusRepository := repositories.NewOrderStatusRepository(client)
	api.OrdersAddNewOrderStatusHandler = orderStatus.AddNewStatusToOrder(statusRepository)
	api.OrdersGetFullOrderHistoryHandler = orderStatus.OrderStatusesHistory(statusRepository)

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
