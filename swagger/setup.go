package swagger

import (
	"net/http"

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

func SetupAPI(entClient *ent.Client, logger *zap.Logger, config *config.AppConfig) (http.Handler, error) {

	passwordGenerator, err := utils.NewPasswordGenerator(config.PasswordConfig.PasswordLength)
	if err != nil {
		return nil, err
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, err
	}
	//repos
	passwordRepo := repositories.NewPasswordResetRepository()
	regConfirmRepo := repositories.NewRegistrationConfirmRepository()
	userRepository := repositories.NewUserRepository()
	tokenRepository := repositories.NewTokenRepository()
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
	handlers.SetActiveAreaHandler(logger, api)
	handlers.SetBlockerHandler(logger, api)
	handlers.SetEquipmentHandler(logger, api, fileManager)
	handlers.SetCategoryHandler(logger, api)
	handlers.SetSubcategoryHandler(logger, api)
	handlers.SetOrderHandler(logger, api)
	handlers.SetOrderStatusHandler(logger, api)
	handlers.SetPasswordResetHandler(logger, api, passwordService)
	handlers.SetPetSizeHandler(logger, api)
	handlers.SetPhotoHandler(logger, api, fileManager)
	handlers.SetRegistrationHandler(logger, api, regConfirmService)
	handlers.SetRoleHandler(logger, api)
	handlers.SetEquipmentStatusNameHandler(logger, api)
	handlers.SetUserHandler(logger, api, tokenManager, regConfirmService)
	handlers.SetPetKindHandler(logger, api)
	return middlewares.Tx(entClient)(api.Serve(nil)), nil
}
