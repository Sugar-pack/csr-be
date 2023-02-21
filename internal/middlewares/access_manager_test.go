package middlewares

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
)

const (
	userRole                  = "user"
	adminRole                 = "admin"
	simpleValidPath           = "/v1/simple"
	validPathWithParam        = "/v1/with/params/{id}"
	validPathWithParamExample = "/v1/with/params/1"
	validPathWithParams       = "/v1/with/params/{id}/and/{name}"
	simpleInvalidPath         = "/v1/invalid"
)

func Test_blackListAccessManager(t *testing.T) {
	var manager AccessManager
	t.Run("NewAccessManager", func(t *testing.T) {
		roles := []string{userRole, adminRole}
		fullAccessRoles := []string{adminRole}
		endpoints := existingEndpoints{
			http.MethodGet: {
				simpleValidPath,
				validPathWithParam,
				validPathWithParams,
			},
		}
		manager = NewAccessManager(roles, fullAccessRoles, endpoints)
	})

	t.Run("AddNewAccess", func(t *testing.T) {
		type accessRule struct {
			role   string
			method string
			path   string
			isErr  bool
			isOk   bool
		}
		newAccessRules := []accessRule{
			{
				role:   userRole,
				method: http.MethodGet,
				path:   simpleValidPath,
				isOk:   true,
			},
			{
				role:   userRole,
				method: http.MethodGet,
				path:   simpleValidPath + "/",
				isErr:  true,
			},
			{
				role:   userRole,
				method: http.MethodGet,
				path:   simpleInvalidPath,
				isErr:  true,
			},
			{
				role:   userRole,
				method: http.MethodGet,
				path:   validPathWithParam,
				isOk:   true,
			},
			{
				role:   userRole,
				method: http.MethodPut,
				path:   validPathWithParam,
				isErr:  true,
			},
			{
				role:   adminRole,
				method: http.MethodGet,
				path:   validPathWithParams,
				isOk:   false,
				isErr:  false,
			},
			{
				role:   userRole,
				method: http.MethodGet,
				path:   simpleInvalidPath,
				isErr:  true,
			},
			{
				role:   "unknown",
				method: http.MethodGet,
				path:   validPathWithParams,
				isErr:  true,
			},
		}
		for _, rule := range newAccessRules {
			ok, err := manager.AddNewAccess(rule.role, rule.method, rule.path)
			if rule.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equalf(t, rule.isOk, ok, "AddNewAccess(%s, %s, %s)", rule.role, rule.method, rule.path)
		}
	})

	type requestData struct {
		role, method, path string
		hasAccess          bool
	}
	requestsData := []requestData{
		{
			role:      userRole,
			method:    http.MethodGet,
			path:      endpointConversion(simpleValidPath),
			hasAccess: true,
		},
		{
			role:      adminRole,
			method:    http.MethodGet,
			path:      endpointConversion(simpleValidPath),
			hasAccess: true,
		},
		{
			role:      userRole,
			method:    http.MethodGet,
			path:      endpointConversion(validPathWithParamExample),
			hasAccess: true,
		},
		{
			role:      userRole,
			method:    http.MethodGet,
			path:      strings.TrimPrefix(endpointConversion(validPathWithParamExample), "/"),
			hasAccess: true,
		},
		{
			role:      userRole,
			method:    http.MethodGet,
			path:      strings.TrimPrefix(endpointConversion(validPathWithParamExample), "/") + "/",
			hasAccess: true,
		},
		{
			role:      userRole,
			method:    http.MethodPut,
			path:      endpointConversion(validPathWithParamExample),
			hasAccess: false,
		},
	}

	t.Run("HasAccess", func(t *testing.T) {
		for _, data := range requestsData {
			assert.Equalf(t, data.hasAccess, manager.HasAccess(data.role, data.method, data.path),
				"HasAccess(%s, %s, %s)", data.role, data.method, data.path)
		}
	})

	t.Run("Authorize", func(t *testing.T) {
		for _, data := range requestsData {
			request := &http.Request{
				Method: data.method,
				URL: &url.URL{
					Path: data.path,
				},
			}
			auth := authentication.Auth{
				Role: &authentication.Role{
					Slug: data.role,
				},
			}
			err := manager.Authorize(request, auth)
			if data.hasAccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		}

	})
}
