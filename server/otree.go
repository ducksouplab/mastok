package server

import (
	"encoding/json"
	"net/http"
)

func getOtreeJSON(path string, target any) error {
	req, _ := http.NewRequest("GET", otreeUrl+path, nil)
	req.Header.Add("otree-rest-key", otreeRestKey)
	r, err := (&http.Client{}).Do(req)

	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
