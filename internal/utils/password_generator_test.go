package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPasswordGenerator_PasswordLengthTooLow(t *testing.T) {
	generator, err := NewPasswordGenerator(1)
	require.Error(t, err)
	require.Nil(t, generator)
}

func TestNewPasswordGenerator_PasswordLengthTooBig(t *testing.T) {
	generator, err := NewPasswordGenerator(33)
	require.Error(t, err)
	require.Nil(t, generator)
}

func TestNewPasswordGenerator_OK(t *testing.T) {
	generator, err := NewPasswordGenerator(10)
	require.NoError(t, err)
	require.NotNil(t, generator)
}

func TestPasswordGenerator_Generate(t *testing.T) {
	length := 10
	generator, err := NewPasswordGenerator(length)
	require.NoError(t, err)
	password, err := generator.NewPassword()
	require.NoError(t, err)
	require.Len(t, password, length)
}
