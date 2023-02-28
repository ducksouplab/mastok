package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvOr_Unit(t *testing.T) {
	t.Run("finds value if environment variable is set", func(t *testing.T) {
		os.Setenv("MASTOK_VAR", "value")
		defer os.Unsetenv("MASTOK_VAR")

		actual := GetEnvOr("MASTOK_VAR", "other")
		assert.Equal(t, "value", actual)
	})

	t.Run("uses default value if environment variable is not set", func(t *testing.T) {
		actual := GetEnvOr("MASTOK_UNSET_VAR", "value")
		assert.Equal(t, "value", actual)
	})
}
