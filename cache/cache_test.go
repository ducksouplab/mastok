package cache

import (
	"testing"

	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestCache_Unit(t *testing.T) {
	t.Run("populates experiments config cache from oTree", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()

		eCache := GetOTreeConfigs()

		assert.Equal(t, "brainstorm", eCache[0].Name)
		assert.Equal(t, "chatroulette", eCache[1].Name)
		assert.Equal(t, "test_config_1_to_8", eCache[2].Name)
	})
}
