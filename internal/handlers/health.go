package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/system"
)

func SetHealthHandler(logger *zap.Logger, api *operations.BeAPI) {
	petKindHandler := NewHealth(logger)

	api.SystemGetHealthHandler = petKindHandler.GetHealthFunc()

}

type Health struct {
	logger *zap.Logger
}

func NewHealth(logger *zap.Logger) *Health {
	return &Health{
		logger: logger,
	}
}

func (pk Health) GetHealthFunc() system.GetHealthHandlerFunc {
	return func(p system.GetHealthParams) middleware.Responder {
		return system.NewGetHealthOK()
	}
}
