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
	waitingLimit       string
}

func newCampaignForm(namespace string) campaignForm {
	return campaignForm{
		namespace:          namespace,
		slug:               namespace + "_slug",
		otreeExperiment:    "xp_name",
		perSession:         "8",
		joinOnce:           "false",
		maxSessions:        "4",
		concurrentSessions: "2",
		sessionDuration:    "15",
		waitingLimit:       "5",
	}
}

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
	data.Set("waiting_limit", cf.waitingLimit)
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
		data := campaignFormData(newCampaignForm("namespace1"))
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
		cf1 := newCampaignForm("namespace2")
		cf2 := newCampaignForm("namespace2")
		cf2.slug = "namespace2_different_slug"
		data := campaignFormData(cf1)
		dataDupNamespace := campaignFormData(cf2)
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
		cf1 := newCampaignForm("namespace3")
		cf2 := newCampaignForm("namespace3bis")
		cf1.slug = "namespace3_slug"
		cf2.slug = "namespace3_slug"
		data := campaignFormData(cf1)
		dataDupSlug := campaignFormData(cf2)
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
		cf := newCampaignForm("namespace4#")
		cf.slug = "namespace4_slug"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if namespace too short", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("n")
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if too many participants per session", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace6")
		cf.perSession = "100"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if missing slug", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		data := campaignFormData(newCampaignForm("namespace7"))
		data.Del("slug")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns", data)
		assert.Equal(t, 422, res.Code)
	})
}
