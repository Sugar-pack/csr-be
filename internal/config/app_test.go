package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppConfig(t *testing.T) {
	cfg, err := GetAppConfig("../..")
	assert.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
}
