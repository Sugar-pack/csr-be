package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	l, err := Get()
	assert.NoError(t, err)
	assert.NotNil(t, l)
}
