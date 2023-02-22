package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// "github.com/jarcoal/httpmock"

	"github.com/stretchr/testify/assert"
)

type jsonItem map[string]any

var sessionsData []jsonItem = []jsonItem{
	{
		"admin_url":        otreeUrl + "/SessionStartLinks/1",
		"code":             "code1",
		"config_name":      "config1",
		"created_at":       1676380785.2318387,
		"label":            "",
		"num_participants": 8,
		"session_wide_url": otreeUrl + "/join/1",
	},
	{
		"admin_url":        otreeUrl + "/SessionStartLinks/2",
		"code":             "code2",
		"config_name":      "config2",
		"created_at":       1676380786.2318387,
		"label":            "",
		"num_participants": 4,
		"session_wide_url": otreeUrl + "/join/2",
	},
}
var sessionDetails1 jsonItem = jsonItem{
	"config": jsonItem{
		"id": "id1",
	},
}

var sessionDetails2 jsonItem = jsonItem{
	"config": jsonItem{
		"id": "id2",
	},
}

func TestSessions(t *testing.T) {
	interceptOtreeGetJSON("/api/sessions", sessionsData)
	interceptOtreeGetJSON("/api/sessions/code1", sessionDetails1)
	interceptOtreeGetJSON("/api/sessions/code2", sessionDetails2)
	defer interceptOff()

	w := testMastokGetRequestWithAuth("/sessions")

	assert.Equal(t, w.Code, 200)
	// presence of "session.config.id"
	assert.Contains(t, w.Body.String(), "id1")
	assert.Contains(t, w.Body.String(), "id2")
	// presence of "session.config_name"
	assert.Contains(t, w.Body.String(), "config1")
	assert.Contains(t, w.Body.String(), "config2")
	// presence of "session.creation date"
	assert.Contains(t, w.Body.String(), "2023")
	// presence of "session.num_participants"
	assert.Contains(t, w.Body.String(), "8")
	assert.Contains(t, w.Body.String(), "4")
	// presence of "session.code"
	assert.Contains(t, w.Body.String(), "code1")
	assert.Contains(t, w.Body.String(), "code2")
	// presence of "session.admin_url"
	assert.Contains(t, w.Body.String(), "SessionStartLinks/1")
	assert.Contains(t, w.Body.String(), "SessionStartLinks/2")
}

func TestSessions_Unauthorized(t *testing.T) {
	interceptOtreeGetJSON("/api/sessions", sessionsData)
	defer interceptOff()

	// incorrect basic auth
	router := NewRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sessions", nil)
	req.Header.Add("Authorization", "Basic "+basicAuth("admin", "incorrect"))
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 401)
}
