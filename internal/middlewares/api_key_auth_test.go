package middlewares

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
)

const tokenString = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MX0.hd6SCn616Yum7hzklWuFQ5okxz_k3TifhWlqbFFugIQ"

// tokenString represents a JWT token that was generated online at https://jwt.io using the payload:
// {
//   "id": 1
// }
// and the secret key "123".

func TestAPIKeyAuthFunc(t *testing.T) {
	ctx := context.TODO()
	userID := 1
	key := "123"
	mockUserRepository := &mocks.UserRepository{}

	expectedUser := &ent.User{
		ID:    userID,
		Edges: ent.UserEdges{Role: &ent.Role{Slug: "user"}},
	}
	mockUserRepository.On("GetUserByID", ctx, userID).Return(expectedUser, nil)

	authFunc := APIKeyAuthFunc(key, mockUserRepository)
	newCtx, principal, err := authFunc(ctx, tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, newCtx)

	principalObj, ok := principal.(*models.Principal)
	assert.True(t, ok, "principal should be of type *models.Principal")
	assert.Equal(t, &models.Principal{
		ID:                      int64(userID),
		Role:                    "user",
		IsRegistrationConfirmed: false,
		IsPersonalDataConfirmed: false,
		IsReadonly:              false,
	}, principalObj)

	mockUserRepository.AssertExpectations(t)
}
