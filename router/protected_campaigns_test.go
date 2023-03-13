package router

import (
	"net/url"
	"testing"

	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func campaignFormData(namespace, slug, config, perSession, maxSessions, concurrentSessions string) url.Values {
	data := url.Values{}
	data.Set("namespace", namespace)
	data.Set("slug", slug)
	data.Set("config", config)
	data.Set("per_session", perSession)
	data.Set("max_sessions", maxSessions)
	data.Set("concurrent_sessions", concurrentSessions)
	return data
}

func TestCampaigns_Templates(t *testing.T) {
	router := getTestRouter()
	t.Run("shows campaigns list", func(t *testing.T) {
		res := MastokGetRequestWithAuth(router, "/campaigns")

		assert.Contains(t, res.Body.String(), "New")
	})

	t.Run("shows campaigns new form", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		res := MastokGetRequestWithAuth(router, "/campaigns/new")

		// t.Log(res.Body.String())
		assert.Equal(t, 200, res.Code)
		assert.Contains(t, res.Body.String(), "Create")
	})
}

func TestCampaigns_Integration(t *testing.T) {
	router := getTestRouter()

	t.Run("creates then lists then supervises", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData("namespace1", "slug1", "config1", "8", "4", "2")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		// t.Log(res.Body.String())
		assert.Equal(t, 302, res.Code)
		// GET list
		res = MastokGetRequestWithAuth(router, "/campaigns")
		assert.Contains(t, res.Body.String(), "namespace1")
		// campaign automatically created with state "Paused"
		assert.Contains(t, res.Body.String(), "Paused")
		// when there is at least one campaign, there should be a Control button
		assert.Contains(t, res.Body.String(), "Supervise")
		// GET supervise
		res = MastokGetRequestWithAuth(router, "/campaigns/supervise/namespace1")
		assert.Contains(t, res.Body.String(), "Supervising")
	})

	t.Run("fails creating if duplicate namespace", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData("namespace2", "slug2", "config1", "8", "4", "2")
		dataDupNamespace := campaignFormData("namespace2", "slug2bis", "config2", "8", "4", "2")
		// POST
		MastokPostRequestWithAuth(router, "/campaigns", data)
		res := MastokPostRequestWithAuth(router, "/campaigns", dataDupNamespace)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if duplicate slug", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData("namespace3", "slug3", "config1", "8", "4", "2")
		dataDupNamespace := campaignFormData("namespace3bis", "slug3", "config2", "8", "4", "2")
		// POST
		MastokPostRequestWithAuth(router, "/campaigns", data)
		res := MastokPostRequestWithAuth(router, "/campaigns", dataDupNamespace)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if invalid character", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData("namespace4#", "slug4", "config1", "8", "4", "2")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if namespace too short", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData("n", "slug5", "config1", "8", "4", "2")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if too many participants per session", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData("namespace6", "slug6", "config1", "100", "4", "2")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if missing slug", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := url.Values{}
		data.Set("namespace", "nsnoslug")
		data.Set("config", "config1")
		data.Set("per_session", "8")
		data.Set("max_sessions", "4")
		data.Set("concurrent_sessions", "2")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})

}
