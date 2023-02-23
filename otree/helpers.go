package otree

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ducksouplab/mastok/config"
)

func GetOTreeJSON(path string, target any) error {
	// request
	req, _ := http.NewRequest(http.MethodGet, config.OTreeURL+path, nil)
	req.Header.Add("otree-rest-key", config.OTreeKey)
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
