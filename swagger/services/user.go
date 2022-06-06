package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

const (
	accessExpireTime  = 15 * time.Minute
	refreshExpireTime = 148 * time.Hour
)

type UserService interface {
	GenerateAccessToken(ctx context.Context, login, password string) (string, bool, error)
}

type userService struct {
	userRepository  repositories.UserRepository
	tokenRepository repositories.TokenRepository
	jwtSecret       string
	logger          *zap.Logger
}

func NewUserService(userRepository repositories.UserRepository, tokenRepository repositories.TokenRepository,
	jwtSecret string, logger *zap.Logger) UserService {
	return &userService{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		jwtSecret:       jwtSecret,
		logger:          logger,
	}
}

// GenerateAccessToken generates access token for user. It returns token string, is it internal error and error.
func (u *userService) GenerateAccessToken(ctx context.Context, login, password string) (string, bool, error) {
	user, err := u.userRepository.GetUserByLogin(ctx, login)
	if ent.IsNotFound(err) {
		return "", false, err
	}
	if err != nil {
		return "", true, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", false, err
	}

	accessToken, err := generateJWT(user, u.jwtSecret)
	if err != nil {
		return "", true, err
	}

	refreshToken, err := generateRefreshToken(user, u.jwtSecret)
	if err != nil {
		return "", true, err
	}

	err = u.tokenRepository.CreateTokens(ctx, user.ID, accessToken, refreshToken)
	if err != nil {
		return "", true, err
	}

	return accessToken, false, nil
}

func generateJWT(user *ent.User, jwtSecretKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["id"] = user.ID
	claims["login"] = user.Login
	claims["role"] = nil
	claims["group"] = nil
	role := user.Edges.Role
	if role == nil {
		return "", errors.New("role is nil")
	}
	claims["role"] = map[string]interface{}{
		"id":   role.ID,
		"slug": role.Slug,
	}

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
