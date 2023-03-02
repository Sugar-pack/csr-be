package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

const (
	accessExpireTime  = 60 * time.Minute
	refreshExpireTime = 148 * time.Hour
)

type tokenManager struct {
	userRepository  domain.UserRepository
	tokenRepository domain.TokenRepository
	jwtSecret       string
	logger          *zap.Logger
}

func (s *tokenManager) RefreshToken(ctx context.Context, token string) (string, string, bool, error) {
	claims := jwt.MapClaims{}
	refreshToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error decoding token")
		}
		return []byte(s.jwtSecret), nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) {
		err = s.tokenRepository.DeleteTokensByRefreshToken(ctx, token)
		if err != nil {
			return "", "", true, err
		}
		return "", "", true, nil
	}

	if err != nil {
		return "", "", true, err
	}

	if refreshToken.Valid {
		if refreshToken.Raw != token {
			return "", "", false, errors.New("refresh token is invalid")
		}

		userID := int(claims["id"].(float64))
		currentUser, errGet := s.userRepository.GetUserByID(ctx, userID) // get current user
		if errGet != nil {
			return "", "", false, errGet
		}

		newAccessToken, errGenJWT := generateJWT(currentUser, s.jwtSecret)
		if errGet != nil {
			s.logger.Error("generate JWT token error")
			return "", "", false, errGenJWT
		}

		newRefreshToken, errGenRefreshToken := generateRefreshToken(currentUser, s.jwtSecret)
		if errGet != nil {
			s.logger.Error("generate refresh token error")
			return "", "", false, errGenRefreshToken
		}

		errDelete := s.tokenRepository.DeleteTokensByRefreshToken(ctx, token)
		if errDelete != nil {
			s.logger.Error("delete tokens error")
			return "", "", true, errDelete
		}

		errCreate := s.tokenRepository.CreateTokens(ctx, userID, newAccessToken, newRefreshToken)
		if errCreate != nil {
			s.logger.Error("create tokens error")
			return "", "", true, errCreate
		}

		return newAccessToken, newRefreshToken, false, nil
	}

	s.logger.Error("token not valid", zap.String("token", token))
	return "", "", true, errors.New("token not valid")
}

func NewTokenManager(userRepository domain.UserRepository, tokenRepository domain.TokenRepository,
	jwtSecret string, logger *zap.Logger) domain.TokenManager {
	return &tokenManager{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		jwtSecret:       jwtSecret,
		logger:          logger,
	}
}

// GenerateTokens generates access token for user. It returns access and refresh token, is it internal error and error.
func (s *tokenManager) GenerateTokens(ctx context.Context, login, password string) (string, string, bool, error) {
	user, err := s.userRepository.GetUserByLogin(ctx, login)
	if ent.IsNotFound(err) {
		return "", "", false, err
	}
	if err != nil {
		return "", "", true, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", false, err
	}

	accessToken, err := generateJWT(user, s.jwtSecret)
	if err != nil {
		return "", "", true, err
	}

	refreshToken, err := generateRefreshToken(user, s.jwtSecret)
	if err != nil {
		return "", "", true, err
	}

	err = s.tokenRepository.CreateTokens(ctx, user.ID, accessToken, refreshToken)
	if err != nil {
		return "", "", true, err
	}

	return accessToken, refreshToken, false, nil
}

const (
	IdClaim            = "id"
	LoginClaim         = "login"
	RoleClaim          = "role"
	SlugClaim          = "slug"
	EmailVerifiedClaim = "emailVerified"
	DataVerifiedClaim  = "dataVerified"
	GroupClaim         = "group"
)

func generateJWT(user *ent.User, jwtSecretKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims[IdClaim] = user.ID
	claims[LoginClaim] = user.Login
	claims[RoleClaim] = nil
	claims[GroupClaim] = nil
	role := user.Edges.Role
	if role == nil {
		return "", errors.New("role is nil")
	}
	claims[RoleClaim] = map[string]interface{}{
		IdClaim:   role.ID,
		SlugClaim: role.Slug,
	}
	claims[EmailVerifiedClaim] = false
	if user.Edges.RegistrationConfirm != nil && len(user.Edges.RegistrationConfirm) > 0 {
		claims[EmailVerifiedClaim] = true
	}
	claims[DataVerifiedClaim] = user.IsConfirmed //TODO: is it right field?

	groups := user.Edges.Groups
	if groups == nil {
		return "", errors.New("groups is nil")
	}
	groupsIDs := make([]int, len(groups))
	for i, group := range groups {
		groupsIDs[i] = group.ID
	}
	claims["group"] = map[string]interface{}{
		"ids": groupsIDs,
	}

	claims["exp"] = time.Now().Add(accessExpireTime).Unix()

	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func generateRefreshToken(user *ent.User, jwtSecretKey string) (string, error) {
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	claims := refreshToken.Claims.(jwt.MapClaims)

	claims["id"] = user.ID
	claims["exp"] = time.Now().Add(refreshExpireTime).Unix()

	signedToken, err := refreshToken.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (s *tokenManager) DeleteTokenPair(ctx context.Context, token string) error {
	return s.tokenRepository.DeleteTokensByRefreshToken(ctx, token)
}
