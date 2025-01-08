package middlewares

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	openApiErrors "github.com/go-openapi/errors"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

const apiPrefix = "/api"

const numRoleVariations = 8

type AccessManager interface {
	AddNewAccess(role Role, method, path string) (bool, error)
	VerifyAccess(role Role, method, path string) error
	Authorize(r *http.Request, i interface{}) error
}

type blackListAccessManager struct {
	endpoints       ExistingEndpoints
	acceptableRoles []Role
	fullAccessRoles []Role
	accessMap       map[Role]map[string][]path
}

type ExistingEndpoints map[string][]string

func (e ExistingEndpoints) Validate() error {
	for method, paths := range e {
		if method != http.MethodGet &&
			method != http.MethodPost &&
			method != http.MethodPut &&
			method != http.MethodDelete &&
			method != http.MethodPatch {
			return fmt.Errorf("method %s is not supported", method)
		}
		for _, p := range paths {
			if !strings.HasPrefix(p, "/") {
				return fmt.Errorf("path %s is not valid", p)
			}
		}
	}
	return nil
}

type Role struct {
	Slug                    string
	IsRegistrationConfirmed bool
	IsPersonalDataConfirmed bool
	IsReadonly              bool
}

// NewAccessManager creates new access manager with admin access to all endpoints.
// All roles can be declared with slug only.
func NewAccessManager(acceptableRoles, fullAccessRoles []Role, endpoints ExistingEndpoints) (AccessManager, error) {
	if err := endpoints.Validate(); err != nil {
		return nil, err
	}
	acceptableRoleVariations := allRoleVariation(acceptableRoles)
	fullAccessRoleVariations := allRoleVariation(fullAccessRoles)
	return &blackListAccessManager{
		endpoints:       endpoints,
		acceptableRoles: acceptableRoleVariations,
		fullAccessRoles: fullAccessRoleVariations,
		accessMap:       make(map[Role]map[string][]path),
	}, nil
}

func allRoleVariation(roles []Role) []Role {
	res := make([]Role, 0, numRoleVariations*len(roles))
	variations := []bool{false, true}

	for _, role := range roles {
		for _, IsRegistrationConfirmed := range variations {
			for _, isPersonalDataConfirmed := range variations {
				for _, isReadonly := range variations {
					res = append(res, Role{
						Slug:                    role.Slug,
						IsRegistrationConfirmed: IsRegistrationConfirmed,
						IsPersonalDataConfirmed: isPersonalDataConfirmed,
						IsReadonly:              isReadonly,
					})
				}
			}
		}
	}

	return res
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

func newPath(endpointPath string) (path, error) {
	if strings.Contains(endpointPath, "{") {
		first := strings.Index(endpointPath, "{")
		second := strings.Index(endpointPath, "}")
		if second < first {
			return path{}, fmt.Errorf("incorrect path %s", endpointPath)
		}
		endpointPath = endpointPath[:first] + "(.*)" + endpointPath[second+1:]
		reg, err := regexp.Compile(endpointPath)
		if err != nil {
			return path{}, err
		}
		return path{
			asRegexp: reg,
		}, nil
	}
	if strings.Contains(endpointPath, "{") {
		return path{}, fmt.Errorf("incorrect path %s", endpointPath)
	}

	return path{
		asString: endpointPath,
	}, nil
}

func endpointConversion(path string) string {
	return apiPrefix + path
}

// AddNewAccess adds new access to the access manager. Returns true if access was added, false if access was not added
func (a *blackListAccessManager) AddNewAccess(role Role, endpointMethod, endpointPath string) (bool, error) {
	if !utils.IsValueInList(role, a.acceptableRoles) {
		return false, fmt.Errorf("role %v is not in the list of acceptable roles", role)
	}
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return false, nil
	}
	endpointMethod = strings.ToUpper(endpointMethod)
	paths, ok := a.endpoints[endpointMethod]
	if !ok {
		return false, fmt.Errorf("method %s is not in the list of supported endpoints", endpointMethod)
	}
	if !utils.IsValueInList(endpointPath, paths) {
		return false, fmt.Errorf("path %s is not in the list of existing endpoints", endpointPath)
	}

	newEndpointPath, err := newPath(endpointConversion(endpointPath))
	if err != nil {
		return false, err
	}

	endpointsByRole, ok2 := a.accessMap[role]
	if !ok2 {
		endpointsByRole = make(map[string][]path)
		a.accessMap[role] = endpointsByRole
	}
	pathToUpdate, ok2 := endpointsByRole[endpointMethod]
	if !ok2 {
		pathToUpdate = []path{
			newEndpointPath,
		}
		endpointsByRole[endpointMethod] = pathToUpdate
		a.accessMap[role] = endpointsByRole
		return true, nil
	}
	if !utils.IsValueInList(newEndpointPath, pathToUpdate) {
		pathToUpdate = append(pathToUpdate, newEndpointPath)
		endpointsByRole[endpointMethod] = pathToUpdate
		a.accessMap[role] = endpointsByRole
		return true, nil
	}
	return false, nil

}

// VerifyAccess checks if role has access to the endpoint.
//
// It returns the error with description (go-openapi/error) or `nil“ if role has an access
func (a *blackListAccessManager) VerifyAccess(role Role, method, path string) error {
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return nil
	}
	allowedPaths, ok := a.accessMap[role][method]
	if !ok {
		return openApiErrors.New(http.StatusForbidden, "user is not authorized")
	}
	for _, allowedPath := range allowedPaths {
		if allowedPath.isMatch(path) {
			return nil
		}
	}
	if !role.IsRegistrationConfirmed {
		return openApiErrors.New(http.StatusForbidden, "user has no confirmed email")
	}
	return openApiErrors.New(http.StatusForbidden, "user is not authorized")
}

func (a *blackListAccessManager) Authorize(r *http.Request, auth interface{}) error {
	principal, ok := auth.(*models.Principal)
	if !ok {
		return openApiErrors.New(http.StatusForbidden, "user is not authorized")
	}

	role := Role{
		Slug:                    principal.Role,
		IsRegistrationConfirmed: principal.IsRegistrationConfirmed,
		IsPersonalDataConfirmed: principal.IsPersonalDataConfirmed,
		IsReadonly:              principal.IsReadonly,
	}

	return a.VerifyAccess(role, r.Method, r.URL.Path)
}
