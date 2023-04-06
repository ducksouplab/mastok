package otree

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ducksouplab/mastok/env"
)

// http://otree.host.com/InitializeParticipant/yxj34dh9
func ParticipantStartURL(participantCode string) string {
	return env.OTreePublicURL + "/InitializeParticipant/" + participantCode
}

func GetOTreeJSON(path string, target any) error {
	log.Printf("[otree] GET request to %v", env.OTreeAPIURL+path)
	// request
	req, _ := http.NewRequest(http.MethodGet, env.OTreeAPIURL+path, nil)
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
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(bodyBytes)

	// request
	req, _ := http.NewRequest(http.MethodPost, env.OTreeAPIURL+path, bodyReader)
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
