package middlewares

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	openApiErrors "github.com/go-openapi/errors"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

const apiPrefix = "/api"

// do not change this values, it is used in swagger spec
const forbiddenMessage = "User is not authorized"
const unconfirmedEmailMessage = "User has no confirmed email"

const quantityOfRoleVariations = 4

type AccessManager interface {
	AddNewAccess(role Role, method, path string) (bool, error)
	HasAccess(role Role, method, path string) bool
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
	IsEmailConfirmed        bool
	IsPersonalDataConfirmed bool
}

// NewAccessManager creates new access manager with admin access to all endpoints.
// All roles can be declared with slug only.
func NewAccessManager(roles, fullAccessRoles []Role, endpoints ExistingEndpoints) (AccessManager, error) {
	err := endpoints.Validate()
	if err != nil {
		return nil, err
	}
	accessMap := make(map[Role]map[string][]path)
	fullAccessRoleVariations := allRoleVariation(fullAccessRoles)
	roleVariations := allRoleVariation(roles)
	return &blackListAccessManager{
		endpoints:       endpoints,
		acceptableRoles: roleVariations,
		fullAccessRoles: fullAccessRoleVariations,
		accessMap:       accessMap,
	}, nil
}

func allRoleVariation(roles []Role) []Role {
	allRoleVariations := make([]Role, 0, quantityOfRoleVariations*len(roles))
	for _, role := range roles {
		allRoleVariations = append(allRoleVariations, Role{
			Slug:                    role.Slug,
			IsEmailConfirmed:        role.IsEmailConfirmed,
			IsPersonalDataConfirmed: role.IsPersonalDataConfirmed,
		})
		allRoleVariations = append(allRoleVariations, Role{
			Slug:                    role.Slug,
			IsEmailConfirmed:        role.IsEmailConfirmed,
			IsPersonalDataConfirmed: !role.IsPersonalDataConfirmed,
		})
		allRoleVariations = append(allRoleVariations, Role{
			Slug:                    role.Slug,
			IsEmailConfirmed:        !role.IsEmailConfirmed,
			IsPersonalDataConfirmed: role.IsPersonalDataConfirmed,
		})
		allRoleVariations = append(allRoleVariations, Role{
			Slug:                    role.Slug,
			IsEmailConfirmed:        !role.IsEmailConfirmed,
			IsPersonalDataConfirmed: !role.IsPersonalDataConfirmed,
		})
	}
	return allRoleVariations
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
		endpointPath = endpointPath[:first] + "(.*)" + endpointPath[second+1:]
		reg, err := regexp.Compile(endpointPath)
		if err != nil {
			return path{}, err
		}
		return path{
			asRegexp: reg,
		}, nil
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

// HasAccess checks if role has access to the endpoint
func (a *blackListAccessManager) HasAccess(role Role, method, path string) bool {
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return true
	}
	allowedPaths, ok := a.accessMap[role][method]
	if !ok {
		return false
	}
	for _, allowedPath := range allowedPaths {
		if allowedPath.isMatch(path) {
			return true
		}
	}
	return false
}

func (a *blackListAccessManager) Authorize(r *http.Request, auth interface{}) error {
	userInfo, ok := auth.(authentication.Auth)
	if !ok {
		return openApiErrors.New(http.StatusForbidden, forbiddenMessage)
	}
	role := Role{
		Slug:                    userInfo.Role.Slug,
		IsEmailConfirmed:        userInfo.IsEmailConfirmed,
		IsPersonalDataConfirmed: userInfo.IsPersonalDataConfirmed,
	}

	if !role.IsEmailConfirmed {
		return openApiErrors.New(http.StatusForbidden, unconfirmedEmailMessage)
	}

	if !a.HasAccess(role, r.Method, r.URL.Path) {
		return openApiErrors.New(http.StatusForbidden, forbiddenMessage)
	}
	return nil
}
