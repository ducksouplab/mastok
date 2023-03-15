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

		eCache := GetExperiments()

		assert.Equal(t, "chatroulette", eCache[0])
		assert.Equal(t, "rawroulette", eCache[1])
	})
}
