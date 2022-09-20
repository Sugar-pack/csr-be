package swagger

import (
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"

	"github.com/go-openapi/loads"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/email"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/handlers"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"
)

func SetupAPI(entClient *ent.Client, logger *zap.Logger, config *config.AppConfig) (*operations.BeAPI, error) {

	passwordGenerator, err := utils.NewPasswordGenerator(config.PasswordConfig.PasswordLength)
	if err != nil {
		return nil, err
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, err
	}
	//repos
	passwordRepo := repositories.NewPasswordResetRepository(entClient)
	regConfirmRepo := repositories.NewRegistrationConfirmRepository(entClient)
	userRepository := repositories.NewUserRepository(entClient)
	tokenRepository := repositories.NewTokenRepository(entClient)
	// config
	passwordTTL := config.PasswordConfig.PasswordTokenTTL
	jwtSecret := config.JWTSecret
	photosFolder := config.PhotoService.PhotosFolder
	// services
	mailSendClient := email.NewSenderSmtp(config.EmailService)
	regConfirmService := services.NewRegistrationConfirmService(mailSendClient, userRepository, regConfirmRepo,
		logger, passwordTTL)
	passwordService := services.NewPasswordResetService(mailSendClient, userRepository, passwordRepo, logger, passwordTTL, passwordGenerator)
	tokenManager := services.NewTokenManager(userRepository, tokenRepository, jwtSecret, logger)
	fileManager := services.NewFileManager(photosFolder, logger)
	// swagger api
	api := operations.NewBeAPI(swaggerSpec)
	api.UseSwaggerUI()
	api.BearerAuth = middlewares.BearerAuthenticateFunc(jwtSecret, logger)
	handlers.SetActiveAreaHandler(entClient, logger, api)
	handlers.SetBlockerHandler(entClient, logger, api)
	handlers.SetEquipmentHandler(entClient, logger, api, fileManager)
	handlers.SetCategoryHandler(entClient, logger, api)
	handlers.SetSubcategoryHandler(entClient, logger, api)
	handlers.SetOrderHandler(entClient, logger, api)
	handlers.SetOrderStatusHandler(entClient, logger, api)
	handlers.SetPasswordResetHandler(logger, api, passwordService)
	handlers.SetPetSizeHandler(entClient, logger, api)
	handlers.SetPhotoHandler(entClient, logger, api, fileManager)
	handlers.SetRegistrationHandler(logger, api, regConfirmService)
	handlers.SetRoleHandler(entClient, logger, api)
	handlers.SetEquipmentStatusHandler(entClient, logger, api)
	handlers.SetUserHandler(entClient, logger, api, tokenManager, regConfirmService)
	handlers.SetPetKindHandler(entClient, logger, api)
	return api, nil
}
