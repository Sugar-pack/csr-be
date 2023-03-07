package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateFileName(t *testing.T) {
	f, err := GenerateFileName()
	require.NoError(t, err)
	require.Len(t, f, 32)
	require.Equal(t, -1, strings.LastIndex(f, "-"))
}
