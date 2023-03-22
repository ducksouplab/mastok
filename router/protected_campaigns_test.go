package router

import (
	"net/url"
	"testing"

	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

type campaignForm struct {
	namespace          string
	slug               string
	otreeExperiment    string
	perSession         string
	joinOnce           string
	maxSessions        string
	concurrentSessions string
	sessionDuration    string
}

// func campaignFormData(namespace, slug, expId, perSession, uniqueParticipants, maxSessions, sessionDuration, concurrentSessions string)
func campaignFormData(cf campaignForm) url.Values {
	data := url.Values{}
	data.Set("namespace", cf.namespace)
	data.Set("slug", cf.slug)
	data.Set("otree_experiment_id", cf.otreeExperiment)
	data.Set("per_session", cf.perSession)
	data.Set("join_once", cf.joinOnce)
	data.Set("max_sessions", cf.maxSessions)
	data.Set("concurrent_sessions", cf.concurrentSessions)
	data.Set("session_duration", cf.sessionDuration)
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

func TestCampaignsSupervise_Unit(t *testing.T) {
	router := getTestRouter()

	t.Run("does not find inexistent campaign", func(t *testing.T) {
		ns := "inexistent_ns"
		res := MastokGetRequestWithAuth(router, "/campaigns/supervise/"+ns)
		assert.Equal(t, 404, res.Result().StatusCode)
	})

	t.Run("shows supervise page with campaign info", func(t *testing.T) {
		ns := "fxt_router_ns1"
		res := MastokGetRequestWithAuth(router, "/campaigns/supervise/"+ns)
		assert.Contains(t, res.Body.String(), ns)
		assert.Contains(t, res.Body.String(), "/join/fxt_router_ns1_slug")
	})
}

func TestCampaigns_Integration(t *testing.T) {
	router := getTestRouter()

	t.Run("creates then lists then supervises", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData(campaignForm{
			namespace:          "namespace1",
			slug:               "namespace1_slug",
			otreeExperiment:    "config1",
			perSession:         "8",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
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
		assert.Contains(t, res.Body.String(), "Current pool")
	})

	t.Run("fails creating if duplicate namespace", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData(campaignForm{
			namespace:          "namespace2",
			slug:               "namespace2_slug",
			otreeExperiment:    "config1",
			perSession:         "8",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
		dataDupNamespace := campaignFormData(campaignForm{
			namespace:          "namespace2",
			slug:               "namespace2_different_slug",
			otreeExperiment:    "config1",
			perSession:         "8",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
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
		data := campaignFormData(campaignForm{
			namespace:          "namespace3",
			slug:               "namespace3_slug",
			otreeExperiment:    "config1",
			perSession:         "8",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
		dataDupSlug := campaignFormData(campaignForm{
			namespace:          "namespace3_different",
			slug:               "namespace3_slug",
			otreeExperiment:    "config1",
			perSession:         "8",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
		// POST
		MastokPostRequestWithAuth(router, "/campaigns", data)
		res := MastokPostRequestWithAuth(router, "/campaigns", dataDupSlug)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if invalid character", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData(campaignForm{
			namespace:          "namespace4#",
			slug:               "namespace4_slug",
			otreeExperiment:    "config1",
			perSession:         "8",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if namespace too short", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData(campaignForm{
			namespace:          "n#",
			slug:               "n_slug",
			otreeExperiment:    "config1",
			perSession:         "8",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if too many participants per session", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData(campaignForm{
			namespace:          "namespace6",
			slug:               "namespace6_slug",
			otreeExperiment:    "config1",
			perSession:         "100",
			joinOnce:           "false",
			maxSessions:        "4",
			concurrentSessions: "2",
			sessionDuration:    "15",
		})
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
		data.Set("otree_experiment_id", "xp1")
		data.Set("per_session", "8")
		data.Set("join_once", "false")
		data.Set("max_sessions", "4")
		data.Set("concurrent_sessions", "2")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})

}
