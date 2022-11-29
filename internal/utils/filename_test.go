package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateFileName(t *testing.T) {
	f, err := GenerateFileName()
	assert.NoError(t, err)
	assert.Len(t, f, 32)
	assert.Equal(t, -1, strings.LastIndex(f, "-"))
}
