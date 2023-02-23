package test_helpers

import (
	"github.com/ducksouplab/mastok/config"
	"github.com/h2non/gock"
)

func InterceptOff() {
	gock.Off()
}

func InterceptOtreeGetJSON(path string, json any) {
	gock.New(config.OTreeURL).
		Get(path).
		Reply(200).
		JSON(json)
}

func InterceptOtreeSessionConfigs() {
	InterceptOtreeGetJSON("/api/session_configs", SESSION_CONFIGS)
}
