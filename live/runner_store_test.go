package live

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunnerStore_Unit(t *testing.T) {
	t.Run("is empty at first", func(t *testing.T) {
		size := getRunnerStoreSize()
		assert.Equal(t, 0, size)
	})

	t.Run("is of size 1 when first client is added", func(t *testing.T) {
		getRunner("fxt_live_ns1")
		defer deleteRunner("fxt_live_ns1")

		size := getRunnerStoreSize()
		assert.Equal(t, 1, size)
	})

	t.Run("is of size 1 when client is added twice", func(t *testing.T) {
		getRunner("fxt_live_ns1")
		getRunner("fxt_live_ns1")
		defer deleteRunner("fxt_live_ns1")

		size := getRunnerStoreSize()
		assert.Equal(t, 1, size)
	})
}