package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDB(t *testing.T) {
	_, _, err := GetDB("host=123")
	assert.ErrorContains(t, err, "failed to ping sql connection: failed to connect")
}
