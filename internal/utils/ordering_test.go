package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetOrderFunc(t *testing.T) {
	f, err := GetOrderFunc("desc", "1")
	require.NoError(t, err)
	require.NotNil(t, f)

	f, err = GetOrderFunc("asc", "2")
	require.NoError(t, err)
	require.NotNil(t, f)

	f, err = GetOrderFunc("something", "2")
	require.Nil(t, f)
	require.ErrorContains(t, err, "wrong value for orderBy: something")
}

func TestIsInOrderFields(t *testing.T) {
	require.True(t, IsValueInList("1", []string{"1", "2", "3"}))
	require.False(t, IsValueInList("4", []string{"1", "2", "3"}))
}
