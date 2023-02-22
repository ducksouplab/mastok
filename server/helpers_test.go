package server

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"

	"github.com/h2non/gock"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func interceptOff() {
	gock.Off()
}

func interceptOtreeGetJSON(path string, json any) {
	gock.New(otreeUrl).
		Get(path).
		Reply(200).
		JSON(json)
}

func testMastokGetRequestWithAuth(path string) (w *httptest.ResponseRecorder) {
	router := NewRouter()
	w = httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Add("Authorization", "Basic "+basicAuth("admin", "admin"))
	router.ServeHTTP(w, req)
	return
}
