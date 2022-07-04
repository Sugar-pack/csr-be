package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/roles"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

type Role struct {
	logger *zap.Logger
}

func NewRole(logger *zap.Logger) *Role {
	return &Role{
		logger: logger,
	}
}

func (r Role) GetRolesFunc(repository repositories.RoleRepository) roles.GetRolesHandlerFunc {
	return func(s roles.GetRolesParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		e, err := repository.GetRoles(ctx)
		if err != nil {
			r.logger.Error("query orders failed")
			return roles.NewGetRolesDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant get all roles"))
		}
		listRoles := models.ListRoles{}
		for _, element := range e {
			id := int64(element.ID)
			listRoles = append(listRoles, &models.Role{
				ID:   &id,
				Name: &element.Name,
				Slug: &element.Slug,
			})
		}
		return roles.NewGetRolesOK().WithPayload(listRoles)
	}
}
