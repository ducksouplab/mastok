package test_helpers

import (
	"log"
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
	log.Printf("[gock] GET " + env.OTreeAPIURL + path)
	gock.New(env.OTreeAPIURL).
		Get(path).
		Reply(200).
		JSON(json)
}

func InterceptOtreeGetPrefixJSON(prefix string, json any) {
	log.Printf("[gock] GET prefix " + env.OTreeAPIURL + prefix)
	gock.New(env.OTreeAPIURL).
		AddMatcher(matchWithPrefix(prefix)).
		Persist(). // can be called multiple times
		Reply(200).
		JSON(json)
}

func InterceptOtreePostJSON(path string, json any) {
	log.Printf("[gock] POST " + env.OTreeAPIURL + path)
	gock.New(env.OTreeAPIURL).
		Post(path).
		Persist(). // can be called multiple times
		Reply(200).
		JSON(json)
}

func InterceptOtreeGetSessionConfigs() {
	InterceptOtreeGetJSON("/api/session_configs", OTREE_GET_SESSION_CONFIGS)
}

func InterceptOtreePostSession() {
	//gock.Observe(gock.DumpRequest)
	InterceptOtreePostJSON("/api/sessions/", OTREE_POST_SESSION)
}

func InterceptOtreeGetSession() {
	InterceptOtreeGetPrefixJSON("/api/sessions/", OTREE_GET_SESSION)
}
