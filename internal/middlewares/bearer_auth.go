package middlewares

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
)

func BearerAuthenticateFunc(key interface{}, _ *zap.Logger) func(string) (interface{}, error) {
	return func(bearerToken string) (interface{}, error) {
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(bearerToken, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New(http.StatusUnauthorized, "error decoding token")
			}
			return []byte(key.(string)), nil
		})
		if err != nil {
			return nil, errors.New(http.StatusUnauthorized, "Failed to parse token")
		}
		if token.Valid {
			login := claims["login"].(string)
			id := int(claims["id"].(float64))
			var rolePointer *authentication.Role = nil
			if claims["role"] != nil {
				role, ok := claims["role"].(map[string]interface{})
				if ok {
					roleId, ok1 := role["id"].(float64)
					slug, ok2 := role["slug"].(string)
					if ok1 && ok2 {
						rolePointer = &authentication.Role{
							Id:   int(roleId),
							Slug: slug,
						}
					}
				}
			}
			return authentication.Auth{
				Id:    id,
				Login: login,
				Role:  rolePointer,
			}, nil
		}
		return nil, errors.New(http.StatusUnauthorized, "Invalid token")
	}
}
