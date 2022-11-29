package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	var a interface{}
	auth := Auth{
		Id:    1,
		Login: "login",
		Role: &Role{
			Id:   2,
			Slug: AdminSlug,
		},
	}

	a = auth

	b, err := IsAdmin(a)
	assert.NoError(t, err)
	assert.Equal(t, true, b)

	b, err = IsManager(a)
	assert.NoError(t, err)
	assert.Equal(t, false, b)

	b, err = IsOperator(a)
	assert.NoError(t, err)
	assert.Equal(t, false, b)

	userID, err := GetUserId(a)
	assert.NoError(t, err)
	assert.Equal(t, 1, userID)

	selectedAuth, err := GetAuth(a)
	assert.NoError(t, err)
	assert.Equal(t, &auth, selectedAuth)
}
