package helpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenvOr_Set(t *testing.T) {
	os.Setenv("MASTOK_VAR", "value")
	defer os.Unsetenv("MASTOK_VAR")

	actual := GetenvOr("MASTOK_VAR", "other")
	assert.Equal(t, actual, "value")
}

func TestGetenvOr_Unset(t *testing.T) {
	actual := GetenvOr("MASTOK_UNSET_VAR", "value")
	assert.Equal(t, actual, "value")
}
