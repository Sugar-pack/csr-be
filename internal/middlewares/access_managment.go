package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type AccessManager struct {
	endpoints       existingEndpoints
	acceptableRoles []string
	fullAccessRoles []string
	accessMap       map[string]existingEndpoints
}

type existingEndpoints map[string][]string

// NewAccessManager creates new access manager with admin access to all endpoints
func NewAccessManager(roles, fullAccessRoles []string, endpoints existingEndpoints) *AccessManager {
	accessMap := make(map[string]existingEndpoints)
	return &AccessManager{
		endpoints:       endpoints,
		acceptableRoles: roles,
		fullAccessRoles: fullAccessRoles,
		accessMap:       accessMap,
	}
}

func endpointConversion(path string) string {
	return fmt.Sprintf("/api%s", path)
}

// AddNewAccess adds new access to the access manager. Returns true if access was added, false if access was not added
func (a *AccessManager) AddNewAccess(role, method, path string) (bool, error) {
	if !utils.IsValueInList(role, a.acceptableRoles) {
		return false, errors.New(fmt.Sprintf("role %s is not in the list of acceptable roles", role))
	}
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return false, nil
	}
	paths, ok := a.endpoints[method]
	if ok {
		if !utils.IsValueInList(path, paths) {
			return false, errors.New(fmt.Sprintf("path %s is not in the list of existing endpoints", path))
		}

		endpointsByRole, ok := a.accessMap[role]
		if !ok {
			endpointsByRole = make(existingEndpoints)
			a.accessMap[role] = endpointsByRole
		}
		pathToUpdate, ok := endpointsByRole[method]
		if !ok {
			pathToUpdate = []string{
				endpointConversion(path),
			}
			endpointsByRole[method] = pathToUpdate
			a.accessMap[role] = endpointsByRole
			return true, nil
		}
		if !utils.IsValueInList(path, pathToUpdate) {
			pathToUpdate = append(pathToUpdate, endpointConversion(path))
			endpointsByRole[method] = pathToUpdate
			a.accessMap[role] = endpointsByRole
			return true, nil
		}
		return false, nil

	}
	return false, errors.New(fmt.Sprintf("method %s is not in the list of existing endpoints", method))
}

// HasAccess checks if role has access to the endpoint
func (a *AccessManager) HasAccess(role, method, path string) bool {
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return true
	}
	paths, ok := a.accessMap[role][method]
	if ok {
		return utils.IsValueInList(path, paths)
	}
	return false
}

func (a *AccessManager) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				next.ServeHTTP(w, r) // token is not required for some endpoints
			}

			role, err := GetRoleFromToken(tokenString)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if !a.HasAccess(role, r.Method, r.URL.Path) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetRoleFromToken(token string) (string, error) { //TODO: optimize
	claims := jwt.MapClaims{}
	jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error decoding token")
		}
		return nil, nil
	}) // ignoring error because it is already checked in the bearer_auth middleware
	roleFromHeader, ok := claims["role"]
	if !ok {
		return "", errors.New("role is not found in the token")
	}
	roleRaw, err := json.Marshal(roleFromHeader)
	if err != nil {
		return "", err
	}
	role := &authentication.Role{}
	err = json.Unmarshal(roleRaw, role)
	if err != nil {
		return "", err
	}

	return role.Slug, nil
}
