package main

import (
	"github.com/go-openapi/loads"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/email"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/handlers"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/overdue"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetupAPI(entClient *ent.Client, lg *zap.Logger, conf *config.AppConfig) (*restapi.Server, domain.OrderOverdueCheckup, error) {
	passwordGenerator, err := utils.NewPasswordGenerator(conf.Password.Length)
	if err != nil {
		return nil, nil, err
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, nil, err
	}
	// repos
	passwordRepo := repositories.NewPasswordResetRepository()
	regConfirmRepo := repositories.NewRegistrationConfirmRepository()
	userRepository := repositories.NewUserRepository()
	tokenRepository := repositories.NewTokenRepository()
	// conf
	passwordTTL := conf.Password.ResetExpirationMinutes
	jwtSecret := conf.JWTSecretKey
	// services
	mailSendClient := email.NewSenderSmtp(conf.Email, email.NewWrapperSmtp(
		conf.Email.ServerHost,
		conf.Email.ServerPort,
		conf.Email.Password,
	))
	regConfirmService := services.NewRegistrationConfirmService(mailSendClient, userRepository, regConfirmRepo,
		lg, passwordTTL)
	passwordService := services.NewPasswordResetService(mailSendClient, userRepository, passwordRepo, lg, passwordTTL, passwordGenerator)
	tokenManager := services.NewTokenManager(userRepository, tokenRepository, jwtSecret, lg)

	// swagger api
	api := operations.NewBeAPI(swaggerSpec)
	api.UseSwaggerUI()
	api.BearerAuth = middlewares.BearerAuthenticateFunc(jwtSecret, lg)
	handlers.SetActiveAreaHandler(lg, api)
	handlers.SetBlockerHandler(lg, api)
	handlers.SetEquipmentHandler(lg, api)
	handlers.SetCategoryHandler(lg, api)
	handlers.SetSubcategoryHandler(lg, api)
	handlers.SetOrderHandler(lg, api)
	orderStatusRepo, orderFilterRepo, equipmentStatusRepo := handlers.SetOrderStatusHandler(lg, api)
	handlers.SetPasswordResetHandler(lg, api, passwordService)
	handlers.SetPetSizeHandler(lg, api)
	handlers.SetPhotoHandler(lg, api)
	handlers.SetRegistrationHandler(lg, api, regConfirmService)
	handlers.SetRoleHandler(lg, api)
	handlers.SetEquipmentStatusNameHandler(lg, api)
	handlers.SetUserHandler(lg, api, tokenManager, regConfirmService)
	handlers.SetPetKindHandler(lg, api)
	handlers.SetHealthHandler(lg, api)

	// run server
	server := restapi.NewServer(api)
	listeners := []string{"http"}

	server.EnabledListeners = listeners
	server.Host = conf.Server.Host
	server.Port = conf.Server.Port
	server.SetHandler(middlewares.Tx(entClient)(api.Serve(nil)))

	return server,
		overdue.NewOverdueCheckup(orderStatusRepo, orderFilterRepo, equipmentStatusRepo, lg),
		nil
}
