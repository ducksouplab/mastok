package router

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func TestBasicAuth_Integration(t *testing.T) {
	router := getTestRouter()
	t.Run("rejects guest user", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		router.ServeHTTP(res, req)

		assert.Equal(t, 401, res.Code)
	})

	t.Run("authorizes user with correct credentials", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Basic "+basicAuth("mastok", "mastok"))
		router.ServeHTTP(res, req)

		assert.Equal(t, 302, res.Code)
		assert.Equal(t, "/dashboard", res.Header()["Location"][0])
	})

	t.Run("rejects user with incorrect credentials", func(t *testing.T) {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Basic "+basicAuth("mastok", "incorrect"))
		router.ServeHTTP(res, req)

		assert.Equal(t, 401, res.Code)
	})
}
