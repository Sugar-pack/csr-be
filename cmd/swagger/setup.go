package main

import (
	"fmt"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/security"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/docs"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/email"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/handlers"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/overdue"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/roles"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetupAPI(entClient *ent.Client, lg *zap.Logger, conf *config.AppConfig) (*restapi.Server, domain.OrderOverdueCheckup, error) {
	passwordGenerator, err := utils.NewPasswordGenerator(conf.Password.Length)
	if err != nil {
		return nil, nil, err
	}

	swaggerSpec, err := loadSwaggerSpec()
	if err != nil {
		return nil, nil, err
	}
	// repos
	passwordRepo := repositories.NewPasswordResetRepository()
	regConfirmRepo := repositories.NewRegistrationConfirmRepository()
	userRepository := repositories.NewUserRepository()
	tokenRepository := repositories.NewTokenRepository()
	emailConfirmRepository := repositories.NewConfirmEmailRepository()

	// conf
	jwtSecret := conf.JWTSecretKey
	// services
	mailSendClient := email.NewSenderSmtp(conf.Email, email.NewWrapperSmtp(
		conf.Email.ServerHost,
		conf.Email.ServerPort,
		conf.Email.Password,
	))
	regConfirmService := services.NewRegistrationConfirmService(mailSendClient, userRepository, regConfirmRepo,
		lg, conf.Email.ConfirmLinkExpiration)
	passwordService := services.NewPasswordResetService(mailSendClient,
		userRepository, passwordRepo, lg, conf.Password.ResetLinkExpiration, passwordGenerator)
	tokenManager := services.NewTokenManager(userRepository, tokenRepository, jwtSecret, lg)
	changeEmailService := services.NewEmailChangeService(
		mailSendClient, userRepository,
		emailConfirmRepository, lg,
	)
	// swagger api
	api := operations.NewBeAPI(swaggerSpec)
	api.UseSwaggerUI()

	api.APIKeyAuthenticator = func(name string, in string, _ security.TokenAuthentication) runtime.Authenticator {
		return security.APIKeyAuthCtx(name, in, middlewares.APIKeyAuthFunc(jwtSecret, userRepository))
	}

	handlers.SetActiveAreaHandler(lg, api)
	handlers.SetEquipmentHandler(lg, api)
	handlers.SetCategoryHandler(lg, api)
	handlers.SetSubcategoryHandler(lg, api)
	handlers.SetOrderHandler(lg, api)
	orderStatusRepo, orderFilterRepo, equipmentStatusRepo := handlers.SetOrderStatusHandler(lg, api)
	handlers.SetPasswordResetHandler(lg, api, passwordService)
	handlers.SetPetSizeHandler(lg, api)
	handlers.SetPhotoHandler(lg, api)
	handlers.SetRegistrationHandler(lg, api, regConfirmService)
	handlers.SetEmailConfirmHandler(lg, api, changeEmailService)
	handlers.SetRoleHandler(lg, api)
	handlers.SetEquipmentStatusNameHandler(lg, api)
	handlers.SetEquipmentStatusHandler(lg, api)
	handlers.SetEquipmentPeriodsHandler(lg, api)
	handlers.SetUserHandler(lg, api, tokenManager, regConfirmService, changeEmailService)
	handlers.SetPetKindHandler(lg, api)
	handlers.SetHealthHandler(lg, api)

	api.Init()
	accessManager, err := AccessManager(api, conf.AccessBindings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create access manager: %w", err)
	}
	api.APIAuthorizer = accessManager
	// run server
	server := restapi.NewServer(api)
	listeners := []string{"http"}

	server.ConfigureAPI()
	server.EnabledListeners = listeners
	server.Host = conf.Server.Host
	server.Port = conf.Server.Port
	server.SetHandler(
		cors.AllowAll().Handler(
			middlewares.Tx(entClient)(api.Serve(nil)),
		),
	)

	return server,
		overdue.NewOverdueCheckup(orderStatusRepo, orderFilterRepo, equipmentStatusRepo, lg),
		nil
}

func loadSwaggerSpec() (*loads.Document, error) {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, err
	}
	// adding unauthorized error to all endpoints
	code, unauthorizedErr := docs.UnauthorizedError()
	docs.AddErrorToSecuredEndpoints(code, unauthorizedErr, swaggerSpec)

	code, forbiddenErr := docs.ForbiddenError()
	docs.AddErrorToSecuredEndpoints(code, forbiddenErr, swaggerSpec)

	raw, err := swaggerSpec.Spec().MarshalJSON()
	if err != nil {
		return nil, err
	}

	swaggerSpec, err = loads.Analyzed(raw, "")
	if err != nil {
		return nil, err
	}
	return swaggerSpec, nil
}

func AccessManager(api *operations.BeAPI, bindings []config.RoleEndpointBinding) (middlewares.AccessManager, error) {
	acceptableRoles := []middlewares.Role{
		{
			Slug: roles.Admin,
		},
		{
			Slug: roles.User,
		},
		{
			Slug: roles.Operator,
		},
		{
			Slug: roles.Manager,
		},
	}
	fullAccessRoles := []middlewares.Role{
		{
			Slug: roles.Admin,
		},
		{
			Slug: roles.Manager,
		},
		{
			Slug: roles.Operator,
		},
	}

	manager, err := middlewares.NewAccessManager(acceptableRoles, fullAccessRoles, api.GetExistingEndpoints())
	if err != nil {
		return nil, err
	}

	for _, binding := range bindings {
		for verb, paths := range binding.AllowedEndpoints {
			for _, path := range paths {
				_, err = manager.AddNewAccess(binding.Role, verb, path)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return manager, nil
}
