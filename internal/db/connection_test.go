package db

import (
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDB(t *testing.T) {
	_, _, err := GetDB(config.DB{Host: "123"})
	require.ErrorContains(t, err, "failed to ping sql connection: failed to connect")
}
