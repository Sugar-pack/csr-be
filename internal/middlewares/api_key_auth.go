package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/golang-jwt/jwt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/services"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func TokenInvalidError() error {
	// make sure that the error message is exactly as the one in the #/definitions/SwaggerError object.
	return errors.New(http.StatusUnauthorized, messages.ErrInvalidToken)
}

func APIKeyAuthFunc(key interface{}, userRepository domain.UserRepository) func(context.Context, string) (context.Context, interface{}, error) {
	return func(ctx context.Context, token string) (context.Context, interface{}, error) {
		claims := jwt.MapClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, TokenInvalidError()
			}
			return []byte(key.(string)), nil
		})
		if err != nil || !parsedToken.Valid {
			return ctx, nil, TokenInvalidError()
		}

		userID, ok := claims[services.UserIDTokenClaim].(float64)
		if !ok {
			return ctx, nil, fmt.Errorf("invalid user ID format")
		}

		user, err := userRepository.GetUserByID(ctx, int(userID))
		if err != nil {
			return ctx, nil, fmt.Errorf("can't get user by ID")
		}

		principal := createPrincipal(user)
		return ctx, principal, nil
	}
}

func createPrincipal(user *ent.User) *models.Principal {
	if user == nil {
		return nil
	}

	principal := &models.Principal{
		ID:                      int64(user.ID),
		Role:                    user.Edges.Role.Slug,
		IsRegistrationConfirmed: user.IsRegistrationConfirmed,
		IsPersonalDataConfirmed: isPersonalDataConfirmed(user),
		IsReadonly:              user.IsReadonly,
	}

	return principal
}

func isPersonalDataConfirmed(user *ent.User) bool {
	return user != nil &&
		user.Name != "" &&
		user.Surname != nil && *user.Surname != "" &&
		user.Phone != nil && *user.Phone != ""
}
