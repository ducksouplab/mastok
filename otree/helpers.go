package otree

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ducksouplab/mastok/env"
)

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
