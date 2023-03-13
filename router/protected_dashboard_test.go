package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDashboard_Templates(t *testing.T) {
	t.Run("shows dashboard", func(t *testing.T) {
		res := MastokGetRequestWithAuth(getTestRouter(), "/dashboard")

		assert.Equal(t, 200, res.Code)
		assert.Contains(t, res.Body.String(), "campaigns")
		assert.Contains(t, res.Body.String(), "sessions")
	})
}
