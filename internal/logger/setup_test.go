package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	l, err := Get()
	require.NoError(t, err)
	require.NotNil(t, l)
}
