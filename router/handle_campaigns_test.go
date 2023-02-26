package router

import (
	"net/url"
	"testing"

	th "github.com/ducksouplab/mastok/test_helpers"
	"github.com/stretchr/testify/assert"
)

func init() {
	// CAUTION: currently DB is not reinitialized after each test, but at a package level
	th.ReinitTestDB()
}

func campaignFormData(namespace, experimentConfig, perSession, sessionMax string) url.Values {
	data := url.Values{}
	data.Set("namespace", namespace)
	data.Set("experiment_config", experimentConfig)
	data.Set("per_session", perSession)
	data.Set("session_max", sessionMax)
	return data
}

func TestCampaigns_ShowList(t *testing.T) {
	res := th.MastokGetRequestWithAuth(NewRouter(), "/campaigns")

	assert.Contains(t, res.Body.String(), "New")
}

func TestCampaigns_ShowNew(t *testing.T) {
	th.InterceptOtreeSessionConfigs()
	defer th.InterceptOff()
	res := th.MastokGetRequestWithAuth(NewRouter(), "/campaigns/new")

	assert.Equal(t, res.Code, 200)
	assert.Contains(t, res.Body.String(), "Create")
}

func TestCampaigns_CreateSuccess_ThenList(t *testing.T) {
	th.InterceptOtreeSessionConfigs()
	defer th.InterceptOff()
	// fill campaign form
	data := campaignFormData("namespace1", "config1", "8", "4")
	// POST
	res := th.MastokPostRequestWithAuth(NewRouter(), "/campaigns", data)
	t.Log(res.Body.String())
	assert.Equal(t, res.Code, 302)
	// GET list
	res = th.MastokGetRequestWithAuth(NewRouter(), "/campaigns")
	assert.Contains(t, res.Body.String(), "namespace1")
	// campaign automatically created with state "Waiting"
	assert.Contains(t, res.Body.String(), "Waiting")
	// when there is at least one campaign, there should be a Control button
	assert.Contains(t, res.Body.String(), "Supervise")
}

func TestCampaigns_CreateFail_Duplicate(t *testing.T) {
	th.InterceptOtreeSessionConfigs()
	defer th.InterceptOff()
	// fill campaign form
	data := campaignFormData("ns1", "config1", "8", "4")
	dataDupNamespace := campaignFormData("ns1", "config2", "8", "4")
	// POST
	th.MastokPostRequestWithAuth(NewRouter(), "/campaigns", data)
	res := th.MastokPostRequestWithAuth(NewRouter(), "/campaigns", dataDupNamespace)
	t.Log(res.Body.String())
	assert.Equal(t, res.Code, 422)
}

func TestCampaigns_CreateFail_InvalidCharacter(t *testing.T) {
	th.InterceptOtreeSessionConfigs()
	defer th.InterceptOff()
	// fill campaign form
	data := campaignFormData("namespace#", "config1", "8", "4")
	// POST
	res := th.MastokPostRequestWithAuth(NewRouter(), "/campaigns", data)
	t.Log(res.Body.String())
	assert.Equal(t, res.Code, 422)
}

func TestCampaigns_CreateFail_TooShort(t *testing.T) {
	th.InterceptOtreeSessionConfigs()
	defer th.InterceptOff()
	// fill campaign form
	data := campaignFormData("n", "config1", "8", "4")
	// POST
	res := th.MastokPostRequestWithAuth(NewRouter(), "/campaigns", data)
	t.Log(res.Body.String())
	assert.Equal(t, res.Code, 422)
}

func TestCampaigns_CreateFail_TooManyParticipants(t *testing.T) {
	th.InterceptOtreeSessionConfigs()
	defer th.InterceptOff()
	// fill campaign form
	data := campaignFormData("namespacemany", "config1", "100", "4")
	// POST
	res := th.MastokPostRequestWithAuth(NewRouter(), "/campaigns", data)
	t.Log(res.Body.String())
	assert.Equal(t, res.Code, 422)
}
