package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	openApiErrors "github.com/go-openapi/errors"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

const apiPrefix = "/api"

type AccessManager interface {
	AddNewAccess(role, method, path string) (bool, error)
	HasAccess(role, method, path string) bool
	Authorize(r *http.Request, i interface{}) error
}

type blackListAccessManager struct {
	endpoints       existingEndpoints
	acceptableRoles []string
	fullAccessRoles []string
	accessMap       map[string]map[string][]path
}

type existingEndpoints map[string][]string

// NewAccessManager creates new access manager with admin access to all endpoints
func NewAccessManager(roles, fullAccessRoles []string, endpoints existingEndpoints) AccessManager {
	accessMap := make(map[string]map[string][]path)
	return &blackListAccessManager{
		endpoints:       endpoints,
		acceptableRoles: roles,
		fullAccessRoles: fullAccessRoles,
		accessMap:       accessMap,
	}
}

type path struct {
	asString string
	asRegexp *regexp.Regexp
}

func (p *path) isMatch(st string) bool {
	endpointPath := normalizePath(st)
	if p.asString != "" {
		return p.asString == endpointPath
	}
	return p.asRegexp.MatchString(endpointPath)
}

func normalizePath(endpointPath string) string {
	if !strings.HasPrefix(endpointPath, "/") {
		endpointPath = "/" + endpointPath
	}
	endpointPath = strings.TrimSuffix(endpointPath, "/")
	return endpointPath
}

func newPath(endpointPath string) path {
	if strings.Contains(endpointPath, "{") {
		first := strings.Index(endpointPath, "{")
		second := strings.Index(endpointPath, "}")
		endpointPath = endpointPath[:first] + "(.*)" + endpointPath[second+1:]
		return path{
			asRegexp: regexp.MustCompile(endpointPath),
		}
	} else {
		return path{
			asString: endpointPath,
		}
	}
}

func endpointConversion(path string) string {
	return apiPrefix + path
}

// AddNewAccess adds new access to the access manager. Returns true if access was added, false if access was not added
func (a *blackListAccessManager) AddNewAccess(role, endpointMethod, endpointPath string) (bool, error) {
	if !utils.IsValueInList(role, a.acceptableRoles) {
		return false, errors.New(fmt.Sprintf("role %s is not in the list of acceptable roles", role))
	}
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return false, nil
	}
	paths, ok := a.endpoints[endpointMethod]
	if ok {
		if !utils.IsValueInList(endpointPath, paths) {
			return false, errors.New(fmt.Sprintf("path %s is not in the list of existing endpoints", endpointPath))
		}

		endpointsByRole, ok2 := a.accessMap[role]
		if !ok2 {
			endpointsByRole = make(map[string][]path)
			a.accessMap[role] = endpointsByRole
		}
		pathToUpdate, ok2 := endpointsByRole[endpointMethod]
		if !ok2 {
			pathToUpdate = []path{
				newPath(endpointConversion(endpointPath)),
			}
			endpointsByRole[endpointMethod] = pathToUpdate
			a.accessMap[role] = endpointsByRole
			return true, nil
		}
		if !utils.IsValueInList(newPath(endpointPath), pathToUpdate) {
			pathToUpdate = append(pathToUpdate, newPath(endpointConversion(endpointPath)))
			endpointsByRole[endpointMethod] = pathToUpdate
			a.accessMap[role] = endpointsByRole
			return true, nil
		}
		return false, nil

	}
	return false, errors.New(fmt.Sprintf("method %s is not in the list of existing endpoints", endpointMethod))
}

// HasAccess checks if role has access to the endpoint
func (a *blackListAccessManager) HasAccess(role, method, path string) bool {
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return true
	}
	paths, ok := a.accessMap[role][method]
	if ok {
		for _, p := range paths {
			if p.isMatch(path) {
				return true
			}
		}
	}
	return false
}

func (a *blackListAccessManager) Authorize(r *http.Request, auth interface{}) error {
	role := auth.(authentication.Auth).Role.Slug
	if !a.HasAccess(role, r.Method, r.URL.Path) {
		return openApiErrors.New(http.StatusUnauthorized, "this user has no access to this endpoint")
	}
	return nil
}
