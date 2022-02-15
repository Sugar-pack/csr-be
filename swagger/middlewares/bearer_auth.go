package middlewares

import (
	"github.com/go-openapi/errors"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

func BearerAuthenticateFunc(key interface{}, _ *zap.Logger) func(string) (interface{}, error) {
	return func(access string) (interface{}, error) {
		_, err := jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.Unauthenticated(access)
			}

			return key, nil
		})
		if err != nil {
			return nil, errors.Unauthenticated(access)
		}

		return true, nil
	}
}
