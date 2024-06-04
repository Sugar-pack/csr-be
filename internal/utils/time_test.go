package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTimeEqualWithHysteresis(t *testing.T) {
	// }
	data := []struct {
		t1       string
		t2       string
		hyst     string
		layout   string
		expected bool
	}{
		{"2006-01-02", "2006-01-03", "23h", time.DateOnly, false},
		{"2006-01-02", "2006-01-03", "24h", time.DateOnly, true},
		{"2006-01-02", "2006-01-03", "25h", time.DateOnly, true},

		{"2006-01-31", "2006-02-01", "23h", time.DateOnly, false},
		{"2006-01-31", "2006-02-01", "25h", time.DateOnly, true},

		{"1999-05-25", "1999-05-25", "0", time.DateOnly, true},
		{"1999-05-25", "1999-05-25", "4h", time.DateOnly, true},
		{"1999-05-25", "1999-05-26", "24h", time.DateOnly, true},
		{"1999-05-25", "1999-05-26", "23h", time.DateOnly, false},
		{"1999-05-25", "1999-05-26", "1h", time.DateOnly, false},

		{"2024-06-04T12:30:45+02:00", "2024-06-04T12:30:45-03:00", "4h59m", time.RFC3339, false},
		{"2024-06-04T12:30:45+02:00", "2024-06-04T12:30:45-03:00", "5h00m", time.RFC3339, true},
		{"2024-06-04T12:30:45+02:00", "2024-06-04T12:30:45-03:00", "5h40m", time.RFC3339, true},
	}

	for _, d := range data {
		t1, err := time.Parse(d.layout, d.t1)
		require.NoError(t, err)

		t2, err := time.Parse(d.layout, d.t2)
		require.NoError(t, err)

		hyst, err := time.ParseDuration(d.hyst)
		require.NoError(t, err)

		assert.Equal(t, IsTimeEqualWithHysteresis(t1, t2, hyst), d.expected)
	}
}
