package router

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/ducksouplab/mastok/models"
	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

type campaignForm struct {
	otreeExperiment    string
	namespace          string
	slug               string
	perSession         string
	joinOnce           string
	maxSessions        string
	concurrentSessions string
	sessionDuration    string
	waitingLimit       string
	grouping           string
	consent            string
}

func newCampaignForm(namespace string) campaignForm {
	return campaignForm{
		otreeExperiment:    "xp_name",
		namespace:          namespace,
		slug:               namespace + "_slug",
		perSession:         "8",
		joinOnce:           "false",
		maxSessions:        "4",
		concurrentSessions: "2",
		sessionDuration:    "15",
		waitingLimit:       "5",
	}
}

func newCampaignFormFromModel(c *models.Campaign) campaignForm {
	return campaignForm{
		otreeExperiment:    c.OtreeExperiment,
		namespace:          c.Namespace,
		slug:               c.Slug,
		perSession:         strconv.Itoa(c.PerSession),
		joinOnce:           strconv.FormatBool(c.JoinOnce),
		maxSessions:        strconv.Itoa(c.MaxSessions),
		concurrentSessions: strconv.Itoa(c.ConcurrentSessions),
		sessionDuration:    strconv.Itoa(c.SessionDuration),
		waitingLimit:       strconv.Itoa(c.WaitingLimit),
	}
}

func campaignFormData(cf campaignForm) url.Values {
	data := url.Values{}
	data.Set("otree_experiment_id", cf.otreeExperiment)
	data.Set("namespace", cf.namespace)
	data.Set("slug", cf.slug)
	data.Set("per_session", cf.perSession)
	data.Set("join_once", cf.joinOnce)
	data.Set("max_sessions", cf.maxSessions)
	data.Set("concurrent_sessions", cf.concurrentSessions)
	data.Set("session_duration", cf.sessionDuration)
	data.Set("waiting_limit", cf.waitingLimit)
	data.Set("grouping", cf.grouping)
	data.Set("consent", cf.consent)
	return data
}

func TestCampaigns_Templates(t *testing.T) {
	router := getTestRouter()
	t.Run("campaigns page", func(t *testing.T) {
		res := MastokGetRequestWithAuth(router, "/campaigns")

		assert.Contains(t, res.Body.String(), "New")
	})

	t.Run("new campaign", func(t *testing.T) {
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

	t.Run("fails creating if duplicate namespace", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf1 := newCampaignForm("namespace1")
		cf2 := newCampaignForm("namespace1")
		cf2.slug = "namespace2_different_slug"
		data := campaignFormData(cf1)
		dataDupNamespace := campaignFormData(cf2)
		// POST
		MastokPostRequestWithAuth(router, "/campaigns/new", data)
		res := MastokPostRequestWithAuth(router, "/campaigns/new", dataDupNamespace)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if duplicate slug", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf1 := newCampaignForm("namespace2")
		cf2 := newCampaignForm("namespace2bis")
		cf1.slug = "namespace3_slug"
		cf2.slug = "namespace3_slug"
		data := campaignFormData(cf1)
		dataDupSlug := campaignFormData(cf2)
		// POST
		MastokPostRequestWithAuth(router, "/campaigns/new", data)
		res := MastokPostRequestWithAuth(router, "/campaigns/new", dataDupSlug)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if invalid character", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace3#")
		cf.slug = "namespace4_slug"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
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
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if missing slug", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace4")
		data := campaignFormData(cf)
		data.Del("slug")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if too many participants per session", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace5")
		cf.perSession = "100"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if Grouping is missing Action", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace6")
		cf.perSession = "4"
		cf.grouping = "What is your gender?\nMale:2\nFemale:3"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if Grouping and PerSession don't match", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace6")
		cf.perSession = "4"
		cf.grouping = "What is your gender?\nMale:2\nFemale:3\nChoose"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("creates if Grouping and PerSession match", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace7")
		cf.perSession = "7"
		cf.grouping = "What is your gender?\nMale:4\nFemale:3\nChoose"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 302, res.Code)
	})

	t.Run("create then lists then supervises then edits", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		ns := "namespace_scenario"
		cf := newCampaignForm(ns)
		cf.consent = "- [ ] checkbox1\n- [ ] checkbox2"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		// t.Log(res.Body.String())
		assert.Equal(t, 302, res.Code)
		// GET list
		res = MastokGetRequestWithAuth(router, "/campaigns")
		assert.Contains(t, res.Body.String(), ns)
		// campaign automatically created with state "Paused"
		assert.Contains(t, res.Body.String(), "Paused")
		// GET supervise
		assert.Contains(t, res.Body.String(), "Supervise")
		assert.Contains(t, res.Body.String(), "/campaigns/supervise/"+ns)
		res = MastokGetRequestWithAuth(router, "/campaigns/supervise/"+ns)
		assert.Contains(t, res.Body.String(), "Waiting room")
		// assert markdown rendering
		assert.Contains(t, res.Body.String(), "<li><input type=\"checkbox\"")
		// GET edit
		assert.Contains(t, res.Body.String(), "Edit campaign")
		assert.Contains(t, res.Body.String(), "/campaigns/edit/"+ns)
		res = MastokGetRequestWithAuth(router, "/campaigns/edit/"+ns)
		assert.Contains(t, res.Body.String(), "- [ ] checkbox1\n- [ ] checkbox2")
	})

	// no need to InterceptOtreeGetSessionConfigs in editing tests since experiment can't be changed
	t.Run("changing namespace or experiment is prevented silently", func(t *testing.T) {
		// fill campaign form
		ns := "fxt_router_ns2_edit"
		forbiddenNs := "fxt_router_ns2_edit_change"
		forbiddenOtreeExperiment := "not_valid"
		campaign, ok := models.GetCampaignByNamespace(ns)
		assert.True(t, ok)
		cf := newCampaignFormFromModel(campaign)
		cf.namespace = forbiddenNs
		cf.otreeExperiment = forbiddenOtreeExperiment
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/edit/"+ns, data)
		assert.Equal(t, 302, res.Code)
		// namespace has not been updated
		res = MastokGetRequestWithAuth(router, "/campaigns/supervise/"+forbiddenNs)
		assert.Equal(t, 404, res.Code)
		// experiment has not been updated
		res = MastokGetRequestWithAuth(router, "/campaigns/supervise/"+ns)
		assert.Equal(t, 200, res.Code)
		assert.Contains(t, res.Body.String(), campaign.OtreeExperiment)
		assert.NotContains(t, res.Body.String(), forbiddenOtreeExperiment)
	})

	t.Run("fails editing if duplicate slug", func(t *testing.T) {
		// fill campaign form
		ns := "fxt_router_ns2_edit"
		campaign, ok := models.GetCampaignByNamespace(ns)
		assert.True(t, ok)
		cf := newCampaignFormFromModel(campaign)
		cf.slug = "fxt_router_ns1_slug"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/edit/"+ns, data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails editing if missing slug", func(t *testing.T) {
		// fill campaign form
		ns := "fxt_router_ns2_edit"
		campaign, ok := models.GetCampaignByNamespace(ns)
		assert.True(t, ok)
		cf := newCampaignFormFromModel(campaign)
		cf.slug = ""
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/edit/"+ns, data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails editing if too many participants per session", func(t *testing.T) {
		// fill campaign form
		ns := "fxt_router_ns2_edit"
		campaign, ok := models.GetCampaignByNamespace(ns)
		assert.True(t, ok)
		cf := newCampaignFormFromModel(campaign)
		cf.perSession = "100"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/edit/"+ns, data)
		assert.Equal(t, 422, res.Code)
	})
}
