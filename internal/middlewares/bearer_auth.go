package middlewares

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/authentication"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/services"
)

func TokenInvalidError() error {
	// make sure that the error message is exactly as the one in the #/definitions/SwaggerError object.
	return errors.New(http.StatusUnauthorized, "Token is invalid")
}

func BearerAuthenticateFunc(key interface{}, _ *zap.Logger) func(string) (interface{}, error) {
	return func(bearerToken string) (interface{}, error) {
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(bearerToken, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, TokenInvalidError()
			}
			return []byte(key.(string)), nil
		})
		if err != nil {
			return nil, TokenInvalidError()
		}
		if token.Valid {
			login := claims[services.LoginClaim].(string)
			id := int(claims[services.IdClaim].(float64))
			var rolePointer *authentication.Role = nil
			if claims[services.RoleClaim] != nil {
				role, ok := claims[services.RoleClaim].(map[string]interface{})
				if ok {
					roleId, ok1 := role[services.IdClaim].(float64)
					slug, ok2 := role[services.SlugClaim].(string)
					if ok1 && ok2 {
						rolePointer = &authentication.Role{
							Id:   int(roleId),
							Slug: slug,
						}
					}
				}
			}
			isEmailConfirmed := false
			if claims[services.EmailVerifiedClaim] != nil {
				isEmailConfirmed = claims[services.EmailVerifiedClaim].(bool)
			}
			isPersonalDataConfirmed := false
			if claims[services.DataVerifiedClaim] != nil {
				isPersonalDataConfirmed = claims[services.DataVerifiedClaim].(bool)
			}
			isReadonly := false
			if claims[services.ReadonlyAccessClaim] != nil {
				isReadonly = claims[services.ReadonlyAccessClaim].(bool)
			}
			return authentication.Auth{
				Id:                      id,
				Login:                   login,
				IsEmailConfirmed:        isEmailConfirmed,
				IsPersonalDataConfirmed: isPersonalDataConfirmed,
				IsReadonly:              isReadonly,
				Role:                    rolePointer,
			}, nil
		}
		return nil, TokenInvalidError()
	}
}
