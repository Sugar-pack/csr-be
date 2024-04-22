package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAppConfig(t *testing.T) {
	cfg, err := GetAppConfig("../..")
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0", cfg.Server.Host)
	require.Equal(t, 8080, cfg.Server.Port)
}
