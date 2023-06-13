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
	otreeConfigName    string
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

func newCampaignFormWithConfig(config, namespace string) campaignForm {
	return campaignForm{
		otreeConfigName:    config,
		namespace:          namespace,
		slug:               namespace + "_slug",
		perSession:         "8",
		joinOnce:           "false",
		maxSessions:        "4",
		concurrentSessions: "2",
		sessionDuration:    "15",
		waitingLimit:       "5",
		consent:            "#Title\ntext\n[accept]Accept[/accept]",
	}
}

func newCampaignForm(namespace string) campaignForm {
	return newCampaignFormWithConfig("test_config_1_to_8", namespace)
}

func newCampaignFormFromModel(c *models.Campaign) campaignForm {
	return campaignForm{
		otreeConfigName:    c.OTreeConfigName,
		namespace:          c.Namespace,
		slug:               c.Slug,
		perSession:         strconv.Itoa(c.PerSession),
		joinOnce:           strconv.FormatBool(c.JoinOnce),
		maxSessions:        strconv.Itoa(c.MaxSessions),
		concurrentSessions: strconv.Itoa(c.ConcurrentSessions),
		sessionDuration:    strconv.Itoa(c.SessionDuration),
		waitingLimit:       strconv.Itoa(c.WaitingLimit),
		consent:            c.Consent,
	}
}

func campaignFormData(cf campaignForm) url.Values {
	data := url.Values{}
	data.Set("otree_config_name", cf.otreeConfigName)
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

func TestCampaigns_Post_Integration(t *testing.T) {
	router := getTestRouter()

	// Namespace tests

	t.Run("fails creating if missing namespace", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace0")
		data := campaignFormData(cf)
		data.Del("namespace")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if duplicate namespace", func(t *testing.T) {
		// fill campaign form
		cf1 := newCampaignForm("namespace1")
		cf2 := newCampaignForm("namespace1")
		cf2.slug = "namespace2_different_slug"
		data := campaignFormData(cf1)
		dataDupNamespace := campaignFormData(cf2)

		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// POST
		MastokPostRequestWithAuth(router, "/campaigns/new", data)
		res := MastokPostRequestWithAuth(router, "/campaigns/new", dataDupNamespace)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if invalid character", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("namespace2#")
		cf.slug = "namespace2_slug"
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

	// Slug tests

	t.Run("fails creating if missing slug", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_slug1")
		data := campaignFormData(cf)
		data.Del("slug")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if duplicate slug", func(t *testing.T) {
		// fill campaign form
		cf1 := newCampaignForm("ns_for_slug2")
		cf2 := newCampaignForm("ns_for_slug2bis")
		cf1.slug = "ns_for_slug_slug"
		cf2.slug = "ns_for_slug_slug"
		data := campaignFormData(cf1)
		dataDupSlug := campaignFormData(cf2)

		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// POST
		MastokPostRequestWithAuth(router, "/campaigns/new", data)
		res := MastokPostRequestWithAuth(router, "/campaigns/new", dataDupSlug)
		// t.Log(res.Body.String())
		assert.Equal(t, 422, res.Code)
	})

	// PerSession tests

	t.Run("fails creating if missing PerSession", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_per_session1")
		data := campaignFormData(cf)
		data.Del("per_session")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if too many participants per session", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_per_session2")
		cf.perSession = "100"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if PerSession not allowed by oTree config", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignFormWithConfig("test_config_4", "ns_for_per_session3")
		cf.perSession = "3"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("creates if PerSession is allowed by oTree config", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignFormWithConfig("test_config_4", "ns_for_per_session4")
		cf.perSession = "4"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 302, res.Code)
	})

	// MaxSessions tests

	t.Run("fails creating if missing MaxSessions", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_max_sessions")
		data := campaignFormData(cf)
		data.Del("max_sessions")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	// ConcurrentSessions tests

	t.Run("fails creating if missing ConcurrentSessions", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_concurrent_sessions")
		data := campaignFormData(cf)
		data.Del("concurrent_sessions")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	// ConcurrentSessions tests

	t.Run("fails creating if missing SessionDuration", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_session_duration")
		data := campaignFormData(cf)
		data.Del("session_duration")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	// Consent tests

	t.Run("fails creating if missing Consent", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_consent1")
		data := campaignFormData(cf)
		data.Del("consent")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	t.Run("fails creating if Consent misses [accept][/accept]", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_consent2")
		data := campaignFormData(cf)
		data.Set("consent", "#Title")
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 422, res.Code)
	})

	// Grouping tests

	t.Run("fails creating if Grouping is missing Action", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_grouping1")
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
		cf := newCampaignForm("ns_for_grouping2")
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
		cf := newCampaignForm("ns_for_grouping3")
		cf.perSession = "7"
		cf.grouping = "What is your gender?\nMale:4\nFemale:3\nChoose"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 302, res.Code)
	})

	t.Run("creates if Grouping is empty", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		cf := newCampaignForm("ns_for_grouping4") // no grouping by default
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
		assert.Equal(t, 302, res.Code)
	})

}

func TestCampaigns_Scenario(t *testing.T) {
	router := getTestRouter()

	t.Run("create then lists then supervises then edits", func(t *testing.T) {
		th.InterceptOtreeGetSessionConfigs()
		defer th.InterceptOff()
		// fill campaign form
		ns := "namespace_scenario"
		cf := newCampaignForm(ns)
		cf.consent = "- [ ] checkbox1\n- [ ] checkbox2\n[accept]Accept[/accept]"
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/new", data)
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
		forbiddenOTreeConfigName := "not_valid"
		campaign, ok := models.GetCampaignByNamespace(ns)
		assert.True(t, ok)
		cf := newCampaignFormFromModel(campaign)
		cf.namespace = forbiddenNs
		cf.otreeConfigName = forbiddenOTreeConfigName
		data := campaignFormData(cf)
		// POST
		res := MastokPostRequestWithAuth(router, "/campaigns/edit/"+ns, data)
		assert.Equal(t, 302, res.Code)
		// namespace has not been updated
		res = MastokGetRequestWithAuth(router, "/campaigns/supervise/"+forbiddenNs)
		assert.Equal(t, 404, res.Code)
		// config has not been updated
		res = MastokGetRequestWithAuth(router, "/campaigns/supervise/"+ns)
		assert.Equal(t, 200, res.Code)
		assert.Contains(t, res.Body.String(), campaign.OTreeConfigName)
		assert.NotContains(t, res.Body.String(), forbiddenOTreeConfigName)
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
