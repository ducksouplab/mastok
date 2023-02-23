package otree

import (
	"testing"

	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestBoot(t *testing.T) {
	th.InterceptOtreeSessionConfigs()
	defer th.InterceptOff()

	eCache := GetExperimentCache()

	assert.Equal(t, eCache[0].Name, "chatroulette")
	assert.Equal(t, eCache[1].Name, "rawroulette")
}
