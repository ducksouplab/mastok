package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuth_ForbidUnauthorized(t *testing.T) {
	router := NewRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 401)
}

func TestBasicAuth_AcceptCorrectCredentials(t *testing.T) {
	router := NewRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Basic "+basicAuth("admin", "admin"))
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 302)
}

func TestBasicAuth_RejectInCorrectCredentials(t *testing.T) {
	router := NewRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Basic "+basicAuth("admin", "incorrect"))
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 401)
}
