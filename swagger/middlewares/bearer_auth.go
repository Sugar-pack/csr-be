package middlewares

import (
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/authentication"
	"github.com/go-openapi/errors"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

func BearerAuthenticateFunc(key interface{}, _ *zap.Logger) func(string) (interface{}, error) {
	return func(bearerToken string) (interface{}, error) {
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(bearerToken, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("error decoding token")
			}
			return []byte(key.(string)), nil
		})
		if err != nil {
			return nil, err
		}
		if token.Valid {
			login := claims["login"].(string)
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
				Login: login,
				Role:  rolePointer,
			}, nil
		}
		return nil, errors.New(0, "Invalid token")
	}
}
