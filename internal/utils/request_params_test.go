package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetParamInt(t *testing.T) {
	a := 5
	assert.Equal(t, 5, GetValueByPointerOrDefaultValue(&a, 6))

	assert.Equal(t, 6, GetValueByPointerOrDefaultValue(nil, 6))
}
