package authentication

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	require.Equal(t, true, b)

	b, err = IsManager(a)
	require.NoError(t, err)
	require.Equal(t, false, b)

	b, err = IsOperator(a)
	require.NoError(t, err)
	require.Equal(t, false, b)

	userID, err := GetUserId(a)
	require.NoError(t, err)
	require.Equal(t, 1, userID)

	selectedAuth, err := GetAuth(a)
	require.NoError(t, err)
	require.Equal(t, &auth, selectedAuth)
}
