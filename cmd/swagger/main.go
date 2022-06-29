package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"

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
	entClient := ent.NewClient(ent.Driver(drv))

	ctx := context.Background()

	// Run the auto migration tool.
	if err := entClient.Schema.Create(ctx, entMigrate.WithDropIndex(true)); err != nil {
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

	passwordLength := getEnv("PASSWORD_LENGTH", "8")
	passwordLengthInt, err := strconv.Atoi(passwordLength)
	if err != nil {
		logger.Fatal("error parsing password length", zap.Error(err))
	}
	passwordGenerator, err := utils.NewPasswordGenerator(passwordLengthInt)
	if err != nil {
		logger.Fatal("error creating password generator", zap.Error(err))
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

	passwordRepo := repositories.NewPasswordResetRepository(entClient)
	regConfirmRepo := repositories.NewRegistrationConfirmRepository(entClient)

	emailSenderWebsiteUrl := getEnv("EMAIL_SENDER_WEBSITE_URL", "https://csr.golangforall.com/")
	if emailSenderWebsiteUrl == "" {
		log.Fatalln("EMAIL_SENDER_WEBSITE_URL not specified")
	}

	mailSendClient := email.NewSenderSmtp(
		emailSenderWebsiteUrl,
		emailSenderServerHost,
		emailSenderServerPort,
		emailSenderPassword,
		emailSenderFromAddress,
		emailSenderFromName)

	petSizeHandler := handlers.NewPetSize(logger)

	equipmentHandler := handlers.NewEquipment(logger)

	petKindHandler := handlers.NewPetKind(logger)

	userHandler := handlers.NewUser(logger)

	roleHandler := handlers.NewRole(logger)

	kindsHandler := handlers.NewKind(logger)

	photosServerURL := getEnv("PHOTOS_SERVER_URL", "http://localhost:8080/")
	photosFolder := getEnv("PHOTOS_FOLDER", "equipments_photos")
	photosHandler := handlers.NewPhoto(photosServerURL, logger)
	fileManager := services.NewFileManager(photosFolder, logger)

	statusHandler := handlers.NewStatus(logger)
	activeAreasHandler := handlers.NewActiveArea(logger)

	userRepository := repositories.NewUserRepository(entClient)
	ttl := time.Duration(passwordResetExpirationMinutesInt) * time.Minute
	passwordService := services.NewPasswordResetService(mailSendClient, userRepository, passwordRepo, logger,
		&ttl, passwordGenerator)
	passwordResetHandler := handlers.NewPasswordReset(
		logger,
		passwordService,
	)

	regConfirmService := services.NewRegistrationConfirmService(mailSendClient, userRepository, regConfirmRepo, logger, &ttl)
	regConfirmHandler := handlers.NewRegistrationConfirmHandler(
		logger,
		regConfirmService,
	)

	blockerHandler := handlers.NewBlocker(logger)

	ordersHandler := handlers.NewOrder(logger)

	orderStatus := handlers.NewOrderStatus(logger)

	api := operations.NewBeAPI(swaggerSpec)
	api.UseSwaggerUI()

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		logger.Error("JWT_SECRET_KEY not specified", zap.Error(err))
	}
	api.BearerAuth = middlewares.BearerAuthenticateFunc(jwtSecretKey, logger)

	tokenRepository := repositories.NewTokenRepository(entClient)
	tokenManager := services.NewTokenManager(userRepository, tokenRepository, jwtSecretKey, logger)
	api.UsersLoginHandler = userHandler.LoginUserFunc(tokenManager)
	api.UsersRefreshHandler = userHandler.Refresh(tokenManager)

	api.UsersPostUserHandler = userHandler.PostUserFunc(userRepository, regConfirmService)
	api.UsersGetCurrentUserHandler = userHandler.GetUserFunc(userRepository)
	api.UsersPatchUserHandler = userHandler.PatchUserFunc(userRepository)
	api.UsersGetUserHandler = userHandler.GetUserById(userRepository)
	api.UsersGetAllUsersHandler = userHandler.GetUsersList(userRepository)
	api.UsersAssignRoleToUserHandler = userHandler.AssignRoleToUserFunc(userRepository)

	blockerRepository := repositories.NewBlockerRepository(entClient)
	api.UsersBlockUserHandler = blockerHandler.BlockUserFunc(blockerRepository)
	api.UsersUnblockUserHandler = blockerHandler.UnblockUserFunc(blockerRepository)

	roleRepository := repositories.NewRoleRepository(entClient)
	api.RolesGetRolesHandler = roleHandler.GetRolesFunc(roleRepository)

	kindRepository := repositories.NewKindRepository(entClient)
	api.KindsCreateNewKindHandler = kindsHandler.CreateNewKindFunc(kindRepository)
	api.KindsGetKindByIDHandler = kindsHandler.GetKindByIDFunc(kindRepository)
	api.KindsDeleteKindHandler = kindsHandler.DeleteKindFunc(kindRepository)
	api.KindsGetAllKindsHandler = kindsHandler.GetAllKindsFunc(kindRepository)
	api.KindsPatchKindHandler = kindsHandler.PatchKindFunc(kindRepository)

	equipmentStatusRepository := repositories.NewEquipmentStatusRepository(entClient)
	api.StatusPostStatusHandler = statusHandler.PostStatusFunc(equipmentStatusRepository)
	api.StatusGetStatusesHandler = statusHandler.GetStatusesFunc(equipmentStatusRepository)
	api.StatusGetStatusHandler = statusHandler.GetStatusFunc(equipmentStatusRepository)
	api.StatusDeleteStatusHandler = statusHandler.DeleteStatusFunc(equipmentStatusRepository)

	equipmentRepository := repositories.NewEquipmentRepository(entClient)
	api.EquipmentCreateNewEquipmentHandler = equipmentHandler.PostEquipmentFunc(equipmentRepository)
	api.EquipmentGetEquipmentHandler = equipmentHandler.GetEquipmentFunc(equipmentRepository)
	api.EquipmentDeleteEquipmentHandler = equipmentHandler.DeleteEquipmentFunc(equipmentRepository, fileManager)
	api.EquipmentGetAllEquipmentHandler = equipmentHandler.ListEquipmentFunc(equipmentRepository)
	api.EquipmentEditEquipmentHandler = equipmentHandler.EditEquipmentFunc(equipmentRepository)
	api.EquipmentFindEquipmentHandler = equipmentHandler.FindEquipmentFunc(equipmentRepository)

	photoRepository := repositories.NewPhotoRepository(entClient)
	api.PhotosCreateNewPhotoHandler = photosHandler.CreateNewPhotoFunc(photoRepository, fileManager)
	api.PhotosGetPhotoHandler = photosHandler.GetPhotoFunc(photoRepository, fileManager)
	api.PhotosDeletePhotoHandler = photosHandler.DeletePhotoFunc(photoRepository, fileManager)
	api.PhotosDownloadPhotoHandler = photosHandler.DownloadPhotoFunc(photoRepository, fileManager)

	activeAreasRepository := repositories.NewActiveAreaRepository(entClient)
	api.ActiveAreasGetAllActiveAreasHandler = activeAreasHandler.GetActiveAreasFunc(activeAreasRepository)

	api.PasswordResetSendLinkByLoginHandler = passwordResetHandler.SendLinkByLoginFunc()
	api.PasswordResetGetPasswordResetLinkHandler = passwordResetHandler.GetPasswordResetLinkFunc()

	api.RegistrationConfirmSendRegistrationConfirmLinkByLoginHandler = regConfirmHandler.SendRegistrationConfirmLinkByLoginFunc()
	api.RegistrationConfirmVerifyRegistrationConfirmTokenHandler = regConfirmHandler.VerifyRegistrationConfirmTokenFunc()

	orderRepository := repositories.NewOrderRepository(entClient)
	api.OrdersGetAllOrdersHandler = ordersHandler.ListOrderFunc(orderRepository)
	api.OrdersCreateOrderHandler = ordersHandler.CreateOrderFunc(orderRepository)
	api.OrdersUpdateOrderHandler = ordersHandler.UpdateOrderFunc(orderRepository)

	orderStatusRepertory := repositories.NewOrderFilter(entClient)
	api.OrdersGetOrdersByStatusHandler = orderStatus.GetOrdersByStatus(orderStatusRepertory)
	api.OrdersGetOrdersByDateAndStatusHandler = orderStatus.GetOrdersByPeriodAndStatus(orderStatusRepertory)

	statusRepository := repositories.NewOrderStatusRepository(entClient)
	api.OrdersAddNewOrderStatusHandler = orderStatus.AddNewStatusToOrder(statusRepository)
	api.OrdersGetFullOrderHistoryHandler = orderStatus.OrderStatusesHistory(statusRepository)

	orderStatusNameRepository := repositories.NewStatusNameRepository(entClient)
	api.OrdersGetAllStatusNamesHandler = orderStatus.GetAllStatusNames(orderStatusNameRepository)

	petSizeRepo := repositories.NewPetSizeRepository(entClient)
	api.PetSizeGetAllPetSizeHandler = petSizeHandler.GetAllPetSizeFunc(petSizeRepo)
	api.PetSizeEditPetSizeHandler = petSizeHandler.UpdatePetSizeByID(petSizeRepo)
	api.PetSizeDeletePetSizeHandler = petSizeHandler.DeletePetSizeByID(petSizeRepo)
	api.PetSizeCreateNewPetSizeHandler = petSizeHandler.CreatePetSizeFunc(petSizeRepo)
	api.PetSizeGetPetSizeHandler = petSizeHandler.GetPetSizeByID(petSizeRepo)

	petKindRepo := repositories.NewPetKindRepository(entClient)
	api.PetKindGetAllPetKindsHandler = petKindHandler.GetAllPetKindFunc(petKindRepo)
	api.PetKindEditPetKindHandler = petKindHandler.UpdatePetKindByID(petKindRepo)
	api.PetKindDeletePetKindHandler = petKindHandler.DeletePetKindByID(petKindRepo)
	api.PetKindCreateNewPetKindHandler = petKindHandler.CreatePetKindFunc(petKindRepo)
	api.PetKindGetPetKindHandler = petKindHandler.GetPetKindsByID(petKindRepo)

	server := restapi.NewServer(api)
	listeners := []string{"http"}

	server.EnabledListeners = listeners
	server.Host = getEnv("SERVER_HOST", "0.0.0.0")
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
