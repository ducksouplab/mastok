package router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/ducksouplab/mastok/helpers"
	"github.com/gin-gonic/gin"
)

func MastokGetRequestWithAuth(router *gin.Engine, path string) (w *httptest.ResponseRecorder) {
	w = httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Add("Authorization", "Basic "+helpers.BasicAuth("mastok", "mastok"))
	router.ServeHTTP(w, req)
	return
}

func MastokPostRequestWithAuth(router *gin.Engine, path string, data url.Values) (w *httptest.ResponseRecorder) {
	w = httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+helpers.BasicAuth("mastok", "mastok"))
	router.ServeHTTP(w, req)
	return
}
