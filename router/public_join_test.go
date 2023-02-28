package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin_GuestAuthorized(t *testing.T) {
	router := NewRouter()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(res, req)

	assert.Equal(t, res.Code, 401)
}
