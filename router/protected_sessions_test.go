package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/helpers"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

type resource map[string]any

var sessionsData = []resource{
	{
		"admin_url":        env.OTreeURL + "/SessionStartLinks/1",
		"code":             "code1",
		"config_name":      "config1",
		"created_at":       1676380785.2318387,
		"label":            "",
		"num_participants": 8,
		"session_wide_url": env.OTreeURL + "/join/1",
	},
	{
		"admin_url":        env.OTreeURL + "/SessionStartLinks/2",
		"code":             "code2",
		"config_name":      "config2",
		"created_at":       1676380786.2318387,
		"label":            "",
		"num_participants": 4,
		"session_wide_url": env.OTreeURL + "/join/2",
	},
}
var sessionDetails1 = resource{
	"config": resource{
		"id": "id1",
	},
}

var sessionDetails2 = resource{
	"config": resource{
		"id": "id2",
	},
}

func TestSessions_Show(t *testing.T) {
	th.InterceptOtreeGetJSON("/api/sessions", sessionsData)
	th.InterceptOtreeGetJSON("/api/sessions/code1", sessionDetails1)
	th.InterceptOtreeGetJSON("/api/sessions/code2", sessionDetails2)
	defer th.InterceptOff()

	res := th.MastokGetRequestWithAuth(getTestRouter(), "/sessions")

	assert.Equal(t, 200, res.Code)
	// presence of "session.config.id"
	assert.Contains(t, res.Body.String(), "id1")
	assert.Contains(t, res.Body.String(), "id2")
	// presence of "session.config_name"
	assert.Contains(t, res.Body.String(), "config1")
	assert.Contains(t, res.Body.String(), "config2")
	// presence of "session.creation date"
	assert.Contains(t, res.Body.String(), "2023")
	// presence of "session.num_participants"
	assert.Contains(t, res.Body.String(), "8")
	assert.Contains(t, res.Body.String(), "4")
	// presence of "session.code"
	assert.Contains(t, res.Body.String(), "code1")
	assert.Contains(t, res.Body.String(), "code2")
	// presence of "session.admin_url"
	assert.Contains(t, res.Body.String(), "SessionStartLinks/1")
	assert.Contains(t, res.Body.String(), "SessionStartLinks/2")
}

func TestSessions_Integration(t *testing.T) {
	t.Run("rejects unauthorized user", func(t *testing.T) {
		th.InterceptOtreeGetJSON("/api/sessions", sessionsData)
		defer th.InterceptOff()

		// incorrect basic auth
		router := getTestRouter()
		res := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/sessions", nil)
		req.Header.Add("Authorization", "Basic "+helpers.BasicAuth("mastok", "incorrect"))
		router.ServeHTTP(res, req)

		assert.Equal(t, 401, res.Code)
	})
}
