package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/logger"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type AccessManager interface {
	AddNewAccess(role, method, path string) (bool, error)
	HasAccess(role, method, path string) bool
	Middleware() func(next http.Handler) http.Handler
}

type BlackListAccessManager struct {
	endpoints       existingEndpoints
	acceptableRoles []string
	fullAccessRoles []string
	accessMap       map[string]map[string][]path
}

type existingEndpoints map[string][]string

// NewAccessManager creates new access manager with admin access to all endpoints
func NewAccessManager(roles, fullAccessRoles []string, endpoints existingEndpoints) AccessManager {
	accessMap := make(map[string]map[string][]path)
	return &BlackListAccessManager{
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

func (p *path) IsMatch(st string) bool {
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
	return fmt.Sprintf("/api%s", path) //TODO: remove hardcode
}

// AddNewAccess adds new access to the access manager. Returns true if access was added, false if access was not added
func (a *BlackListAccessManager) AddNewAccess(role, endpointMethod, endpointPath string) (bool, error) {
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

		endpointsByRole, ok := a.accessMap[role]
		if !ok {
			endpointsByRole = make(map[string][]path)
			a.accessMap[role] = endpointsByRole
		}
		pathToUpdate, ok := endpointsByRole[endpointMethod]
		if !ok {
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
func (a *BlackListAccessManager) HasAccess(role, method, path string) bool {
	if utils.IsValueInList(role, a.fullAccessRoles) {
		return true
	}
	paths, ok := a.accessMap[role][method]
	if ok {
		for _, p := range paths {
			if p.IsMatch(path) {
				return true
			}
		}
	}
	return false
}

func (a *BlackListAccessManager) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			log, err := logger.Get()
			if err != nil {
				w.Write([]byte("error getting logger"))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Info(fmt.Sprintf("token %s", tokenString))
			if tokenString == "" {
				log.Info("token is not provided")
				next.ServeHTTP(w, r) // token is not required for some endpoints
				return
			}
			role, err := GetRoleFromToken(tokenString)
			if err != nil {
				log.Error(fmt.Sprintf("error getting role from token: %s. token %s", err.Error(), tokenString))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if !a.HasAccess(role, r.Method, r.URL.Path) {
				log.Error(fmt.Sprintf("role %s has no access to %s %s", role, r.Method, r.URL.Path))
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
