package db

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDB(t *testing.T) {
	_, _, err := GetDB("host=123")
	require.ErrorContains(t, err, "failed to ping sql connection: failed to connect")
}
