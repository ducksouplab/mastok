package test_helpers

import (
	"net/http"
	"strings"

	"github.com/ducksouplab/mastok/env"
	"github.com/h2non/gock"
)

func matchWithPrefix(prefix string) gock.MatchFunc {
	return func(req *http.Request, rg *gock.Request) (bool, error) {
		if strings.HasPrefix(req.URL.Path, prefix) {
			return true, nil
		}
		return false, nil
	}
}

func InterceptOff() {
	gock.Off()
}

func InterceptOtreeGetJSON(path string, json any) {
	gock.New(env.OTreeURL).
		Get(path).
		Reply(200).
		JSON(json)
}

func InterceptOtreeGetPrefixJSON(prefix string, json any) {
	gock.New(env.OTreeURL).
		AddMatcher(matchWithPrefix(prefix)).
		Reply(200).
		JSON(json)
}

func InterceptOtreePostJSON(path string, json any) {
	gock.New(env.OTreeURL).
		Post(path).
		Reply(200).
		JSON(json)
}

func InterceptOtreeGetSessionConfigs() {
	InterceptOtreeGetJSON("/api/session_configs", OTREE_GET_SESSION_CONFIGS)
}

func InterceptOtreePostSession() {
	InterceptOtreePostJSON("/api/sessions/", OTREE_POST_SESSION)
}

func InterceptOtreeGetSession() {
	InterceptOtreeGetPrefixJSON("/api/sessions/", OTREE_GET_SESSION)
}
