package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/go-openapi/loads"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	entMigrate "git.epam.com/epm-lstr/epm-lstr-lc/be/ent/migrate"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/email"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/handlers"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"
)

func main() {
	var loggerConfig = zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(zap.DebugLevel)

	logger, err := loggerConfig.Build()
	if err != nil {
		logger.Fatal("load config error", zap.Error(err))
	}

	dbHost := getEnv("DB_HOST", "localhost")

	connectionString := fmt.Sprintf("host=%s user=csr password=csr dbname=csr sslmode=disable", dbHost)
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.Fatal("cant open db", zap.Error(err))
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	ctx := context.Background()

	// Run the auto migration tool.
	if err := client.Schema.Create(ctx, entMigrate.WithDropIndex(true)); err != nil {
		logger.Fatal("failed creating schema resources", zap.Error(err))
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"csr", driver)
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			logger.Fatal("migration failed", zap.Error(err))
		}
		logger.Error("migration error", zap.Error(err))
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		logger.Error("error loading swagger spec", zap.Error(err))
		return
	}

	emailSenderServerHost := os.Getenv("EMAIL_SENDER_SERVER_HOST")
	if emailSenderServerHost == "" {
		log.Fatalln("EMAIL_SENDER_SERVER_HOST not specified")
	}
	emailSenderServerPort := os.Getenv("EMAIL_SENDER_SERVER_PORT")
	if emailSenderServerPort == "" {
		log.Fatalln("EMAIL_SENDER_SERVER_PORT not specified")
	}
	emailSenderPassword := os.Getenv("EMAIL_SENDER_PASSWORD")
	if emailSenderPassword == "" {
		log.Fatalln("EMAIL_SENDER_PASSWORD not specified")
	}
	emailSenderFromAddress := os.Getenv("EMAIL_SENDER_FROM_ADDRESS")
	if emailSenderFromAddress == "" {
		log.Fatalln("EMAIL_SENDER_FROM_ADDRESS not specified")
	}
	emailSenderFromName := os.Getenv("EMAIL_SENDER_FROM_NAME")
	if emailSenderFromName == "" {
		log.Fatalln("EMAIL_SENDER_FROM_NAME not specified")
	}

	passwordResetExpirationMinutes := os.Getenv("PASSWORD_RESET_EXPIRATION_MINUTES")
	if passwordResetExpirationMinutes == "" {
		log.Fatalln("PASSWORD_RESET_EXPIRATION_MINUTES not specified")
	}
	passwordResetExpirationMinutesInt, err := strconv.Atoi(passwordResetExpirationMinutes)
	if err != nil {
		log.Fatalln("PASSWORD_RESET_EXPIRATION_MINUTES not a number")
	}

	passwordRepo := repositories.NewPasswordResetRepository(client)

	host := getEnv("SERVER_HOST", "127.0.0.1")
	if host == "" {
		log.Fatalln("HOST not specified")
	}

	mailSendClient := email.NewSenderSmtp(
		host,
		emailSenderServerHost,
		emailSenderServerPort,
		emailSenderPassword,
		emailSenderFromAddress,
		emailSenderFromName)

	equipmentHandler := handlers.NewEquipment(logger)

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

	userRepository := repositories.NewUserRepository(client)
	ttl := time.Duration(passwordResetExpirationMinutesInt) * time.Minute
	passwordService := services.NewPasswordResetService(mailSendClient, userRepository, passwordRepo, logger, &ttl)
	passwordResetHandler := handlers.NewPasswordReset(
		logger,
		passwordService,
	)

	blockerHandler := handlers.NewBlocker(logger)

	ordersHandler := handlers.NewOrder(
		client,
		logger,
	)

	orderStatus := handlers.NewOrderStatus(logger)

	api := operations.NewBeAPI(swaggerSpec)
	api.UseSwaggerUI()
	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		logger.Error("JWT_SECRET_KEY not specified", zap.Error(err))
	}
	api.BearerAuth = middlewares.BearerAuthenticateFunc(jwtSecretKey, logger)
	api.UsersRefreshHandler = userHandler.Refresh(jwtSecretKey)

	tokenRepository := repositories.NewTokenRepository(client)
	userService := services.NewUserService(userRepository, tokenRepository, jwtSecretKey, logger)
	api.UsersLoginHandler = userHandler.LoginUserFunc(userService)

	api.UsersPostUserHandler = userHandler.PostUserFunc(userRepository)
	api.UsersGetCurrentUserHandler = userHandler.GetUserFunc()
	api.UsersPatchUserHandler = userHandler.PatchUserFunc()
	api.UsersGetUserHandler = userHandler.GetUserById()
	api.UsersGetAllUsersHandler = userHandler.GetUsersList()
	api.UsersBlockUserHandler = blockerHandler.BlockUserFunc(repositories.NewBlockerRepository(client))
	api.UsersUnblockUserHandler = blockerHandler.UnblockUserFunc(repositories.NewBlockerRepository(client))
	api.UsersAssignRoleToUserHandler = userHandler.AssignRoleToUserFunc(userRepository)

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

	equipmentRepository := repositories.NewEquipmentRepository(client)
	api.EquipmentCreateNewEquipmentHandler = equipmentHandler.PostEquipmentFunc(equipmentRepository)
	api.EquipmentGetEquipmentHandler = equipmentHandler.GetEquipmentFunc(equipmentRepository)
	api.EquipmentDeleteEquipmentHandler = equipmentHandler.DeleteEquipmentFunc(equipmentRepository)
	api.EquipmentGetAllEquipmentHandler = equipmentHandler.ListEquipmentFunc(equipmentRepository)
	api.EquipmentEditEquipmentHandler = equipmentHandler.EditEquipmentFunc(equipmentRepository)
	api.EquipmentFindEquipmentHandler = equipmentHandler.FindEquipmentFunc(equipmentRepository)

	api.ActiveAreasGetAllActiveAreasHandler = activeAreasHandler.GetActiveAreasFunc()

	api.PasswordResetSendLinkByLoginHandler = passwordResetHandler.SendLinkByLoginFunc()
	api.PasswordResetGetPasswordResetLinkHandler = passwordResetHandler.GetPasswordResetLinkFunc()

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

	orderStatusNameRepository := repositories.NewStatusNameRepository(client)
	api.OrdersGetAllStatusNamesHandler = orderStatus.GetAllStatusNames(orderStatusNameRepository)

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
