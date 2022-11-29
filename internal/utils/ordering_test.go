package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrderFunc(t *testing.T) {
	f, err := GetOrderFunc("desc", "1")
	assert.NoError(t, err)
	assert.NotNil(t, f)

	f, err = GetOrderFunc("asc", "2")
	assert.NoError(t, err)
	assert.NotNil(t, f)

	f, err = GetOrderFunc("something", "2")
	assert.Nil(t, f)
	assert.ErrorContains(t, err, "wrong value for orderBy: something")
}

func TestIsInOrderFields(t *testing.T) {
	assert.True(t, IsValueInList("1", []string{"1", "2", "3"}))
	assert.False(t, IsValueInList("4", []string{"1", "2", "3"}))
}
