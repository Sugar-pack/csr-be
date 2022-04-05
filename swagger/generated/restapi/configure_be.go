// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/active_areas"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/kinds"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/roles"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/status"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/users"
)

//go:generate swagger generate server --target ../../../../be --name Be --spec ../../spec.yaml --model-package swagger/generated/models --server-package swagger/generated/restapi --principal interface{} --exclude-main

func configureFlags(api *operations.BeAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.BeAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Authorization" header is set
	if api.BearerAuth == nil {
		api.BearerAuth = func(token string) (interface{}, error) {
			return nil, errors.NotImplemented("api key auth (Bearer) Authorization from header param [Authorization] has not yet been implemented")
		}
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	if api.UsersAssignRoleToUserHandler == nil {
		api.UsersAssignRoleToUserHandler = users.AssignRoleToUserHandlerFunc(func(params users.AssignRoleToUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.AssignRoleToUser has not yet been implemented")
		})
	}
	if api.EquipmentCreateNewEquipmentHandler == nil {
		api.EquipmentCreateNewEquipmentHandler = equipment.CreateNewEquipmentHandlerFunc(func(params equipment.CreateNewEquipmentParams) middleware.Responder {
			return middleware.NotImplemented("operation equipment.CreateNewEquipment has not yet been implemented")
		})
	}
	if api.KindsCreateNewKindHandler == nil {
		api.KindsCreateNewKindHandler = kinds.CreateNewKindHandlerFunc(func(params kinds.CreateNewKindParams) middleware.Responder {
			return middleware.NotImplemented("operation kinds.CreateNewKind has not yet been implemented")
		})
	}
	if api.EquipmentDeleteEquipmentHandler == nil {
		api.EquipmentDeleteEquipmentHandler = equipment.DeleteEquipmentHandlerFunc(func(params equipment.DeleteEquipmentParams) middleware.Responder {
			return middleware.NotImplemented("operation equipment.DeleteEquipment has not yet been implemented")
		})
	}
	if api.StatusDeleteStatusHandler == nil {
		api.StatusDeleteStatusHandler = status.DeleteStatusHandlerFunc(func(params status.DeleteStatusParams) middleware.Responder {
			return middleware.NotImplemented("operation status.DeleteStatus has not yet been implemented")
		})
	}
	if api.EquipmentEditEquipmentHandler == nil {
		api.EquipmentEditEquipmentHandler = equipment.EditEquipmentHandlerFunc(func(params equipment.EditEquipmentParams) middleware.Responder {
			return middleware.NotImplemented("operation equipment.EditEquipment has not yet been implemented")
		})
	}
	if api.EquipmentFindEquipmentHandler == nil {
		api.EquipmentFindEquipmentHandler = equipment.FindEquipmentHandlerFunc(func(params equipment.FindEquipmentParams) middleware.Responder {
			return middleware.NotImplemented("operation equipment.FindEquipment has not yet been implemented")
		})
	}
	if api.ActiveAreasGetAllActiveAreasHandler == nil {
		api.ActiveAreasGetAllActiveAreasHandler = active_areas.GetAllActiveAreasHandlerFunc(func(params active_areas.GetAllActiveAreasParams) middleware.Responder {
			return middleware.NotImplemented("operation active_areas.GetAllActiveAreas has not yet been implemented")
		})
	}
	if api.EquipmentGetAllEquipmentHandler == nil {
		api.EquipmentGetAllEquipmentHandler = equipment.GetAllEquipmentHandlerFunc(func(params equipment.GetAllEquipmentParams) middleware.Responder {
			return middleware.NotImplemented("operation equipment.GetAllEquipment has not yet been implemented")
		})
	}
	if api.KindsGetAllKindsHandler == nil {
		api.KindsGetAllKindsHandler = kinds.GetAllKindsHandlerFunc(func(params kinds.GetAllKindsParams) middleware.Responder {
			return middleware.NotImplemented("operation kinds.GetAllKinds has not yet been implemented")
		})
	}
	if api.EquipmentGetEquipmentHandler == nil {
		api.EquipmentGetEquipmentHandler = equipment.GetEquipmentHandlerFunc(func(params equipment.GetEquipmentParams) middleware.Responder {
			return middleware.NotImplemented("operation equipment.GetEquipment has not yet been implemented")
		})
	}
	if api.RolesGetRolesHandler == nil {
		api.RolesGetRolesHandler = roles.GetRolesHandlerFunc(func(params roles.GetRolesParams) middleware.Responder {
			return middleware.NotImplemented("operation roles.GetRoles has not yet been implemented")
		})
	}
	if api.StatusGetStatusHandler == nil {
		api.StatusGetStatusHandler = status.GetStatusHandlerFunc(func(params status.GetStatusParams) middleware.Responder {
			return middleware.NotImplemented("operation status.GetStatus has not yet been implemented")
		})
	}
	if api.StatusGetStatusesHandler == nil {
		api.StatusGetStatusesHandler = status.GetStatusesHandlerFunc(func(params status.GetStatusesParams) middleware.Responder {
			return middleware.NotImplemented("operation status.GetStatuses has not yet been implemented")
		})
	}
	if api.UsersLoginHandler == nil {
		api.UsersLoginHandler = users.LoginHandlerFunc(func(params users.LoginParams) middleware.Responder {
			return middleware.NotImplemented("operation users.Login has not yet been implemented")
		})
	}
	if api.KindsPatchKindHandler == nil {
		api.KindsPatchKindHandler = kinds.PatchKindHandlerFunc(func(params kinds.PatchKindParams) middleware.Responder {
			return middleware.NotImplemented("operation kinds.PatchKind has not yet been implemented")
		})
	}
	if api.KindsDeleteKindHandler == nil {
		api.KindsDeleteKindHandler = kinds.DeleteKindHandlerFunc(func(params kinds.DeleteKindParams) middleware.Responder {
			return middleware.NotImplemented("operation kinds.DeleteKind has not yet been implemented")
		})
	}
	if api.UsersGetCurrentUserHandler == nil {
		api.UsersGetCurrentUserHandler = users.GetCurrentUserHandlerFunc(func(params users.GetCurrentUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetCurrentUser has not yet been implemented")
		})
	}
	if api.KindsGetKindByIDHandler == nil {
		api.KindsGetKindByIDHandler = kinds.GetKindByIDHandlerFunc(func(params kinds.GetKindByIDParams) middleware.Responder {
			return middleware.NotImplemented("operation kinds.GetKindByID has not yet been implemented")
		})
	}
	if api.UsersPatchUserHandler == nil {
		api.UsersPatchUserHandler = users.PatchUserHandlerFunc(func(params users.PatchUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.PatchUser has not yet been implemented")
		})
	}
	if api.StatusPostStatusHandler == nil {
		api.StatusPostStatusHandler = status.PostStatusHandlerFunc(func(params status.PostStatusParams) middleware.Responder {
			return middleware.NotImplemented("operation status.PostStatus has not yet been implemented")
		})
	}
	if api.UsersPostUserHandler == nil {
		api.UsersPostUserHandler = users.PostUserHandlerFunc(func(params users.PostUserParams) middleware.Responder {
			return middleware.NotImplemented("operation users.PostUser has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
