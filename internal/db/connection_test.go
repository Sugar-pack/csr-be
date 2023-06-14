package db

import (
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"

	"github.com/stretchr/testify/require"
)

func TestGetDB(t *testing.T) {
	_, _, err := GetDB(config.DB{Host: "localhost"})
	require.ErrorContains(t, err, "failed to ping sql connection: failed to connect")
}
