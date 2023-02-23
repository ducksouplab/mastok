package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvOr_Set(t *testing.T) {
	os.Setenv("MASTOK_VAR", "value")
	defer os.Unsetenv("MASTOK_VAR")

	actual := GetEnvOr("MASTOK_VAR", "other")
	assert.Equal(t, actual, "value")
}

func TestGetEnvOr_Unset(t *testing.T) {
	actual := GetEnvOr("MASTOK_UNSET_VAR", "value")
	assert.Equal(t, actual, "value")
}
