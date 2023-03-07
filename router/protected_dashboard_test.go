package router

import (
	"testing"

	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func TestDashboard_Templates(t *testing.T) {
	t.Run("shows dashboard", func(t *testing.T) {
		res := th.MastokGetRequestWithAuth(getTestRouter(), "/dashboard")

		assert.Equal(t, 200, res.Code)
		assert.Contains(t, res.Body.String(), "campaigns")
		assert.Contains(t, res.Body.String(), "sessions")
	})
}
