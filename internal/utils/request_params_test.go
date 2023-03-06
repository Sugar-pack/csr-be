package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetParamInt(t *testing.T) {
	a := 5
	require.Equal(t, 5, GetValueByPointerOrDefaultValue(&a, 6))

	require.Equal(t, 6, GetValueByPointerOrDefaultValue(nil, 6))
}
