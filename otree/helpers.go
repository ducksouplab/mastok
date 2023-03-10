package otree

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ducksouplab/mastok/env"
)

// http://localhost:8180/InitializeParticipant/yxj34dh9
func ParticipantStartURL(participantCode string) string {
	return env.OTreeURL + "/InitializeParticipant/" + participantCode
}

func GetOTreeJSON(path string, target any) error {
	// request
	req, _ := http.NewRequest(http.MethodGet, env.OTreeURL+path, nil)
	req.Header.Add("otree-rest-key", env.OTreeKey)
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return errors.New(res.Status + " from oTree")
	}
	// unmarshal
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(target)
}

func PostOTreeJSON(path string, body, target any) error {
	// prepare body
	bodyString, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(bodyString)
	// request
	req, _ := http.NewRequest(http.MethodPost, env.OTreeURL+path, bodyReader)
	req.Header.Add("otree-rest-key", env.OTreeKey)
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return errors.New(res.Status + " from oTree")
	}
	// unmarshal
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(target)
}
