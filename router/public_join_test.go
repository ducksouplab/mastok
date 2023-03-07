package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin_Integration(t *testing.T) {
	t.Run("accepts guest user on public page", func(t *testing.T) {
		router := getTestRouter()

		res := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/join/slug", nil)
		router.ServeHTTP(res, req)

		assert.Equal(t, 200, res.Code)
	})
}
