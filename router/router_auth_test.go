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

func TestBasicAuth_ForbidUnauthorized(t *testing.T) {
	router := NewRouter()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(res, req)

	assert.Equal(t, res.Code, 401)
}

func TestBasicAuth_AcceptCorrectCredentials(t *testing.T) {
	router := NewRouter()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Basic "+basicAuth("mastok", "mastok"))
	router.ServeHTTP(res, req)

	assert.Equal(t, res.Code, 302)
	assert.Equal(t, res.Header()["Location"][0], "/dashboard")
}

func TestBasicAuth_RejectIncorrectCredentials(t *testing.T) {
	router := NewRouter()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Authorization", "Basic "+basicAuth("mastok", "incorrect"))
	router.ServeHTTP(res, req)

	assert.Equal(t, res.Code, 401)
}
