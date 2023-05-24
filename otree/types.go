package otree

import "time"

type ExperimentConfig struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Doc  string `json:"doc"`
}

type NestedParticipant struct {
	Code string `json:"code"`
}

type NestedConfig struct {
	Id string `json:"id"`
}

type Session struct {
	Id              string              `json:"id"`
	Code            string              `json:"code"`
	ConfigName      string              `json:"config_name"`
	CreatedAtFloat  float32             `json:"created_at"`
	NumParticipants int                 `json:"num_participants"`
	AdminUrl        string              `json:"admin_url"`
	Config          NestedConfig        `json:"config"`
	Participants    []NestedParticipant `json:"participants"`
}

type NestedSessionDetails struct {
	Config struct {
		Id string `json:"id"`
	} `json:"config"`
}

// to create session on oTree API
type SessionArgs struct {
	SessionConfigName string       `json:"session_config_name"`
	NumParticipants   int          `json:"num_participants"`
	Config            NestedConfig `json:"modified_session_config_fields"`
}

func (s Session) FormatCreatedAt() string {
	return time.Unix(int64(s.CreatedAtFloat), 0).UTC().Format("2006-01-02 15:04:05")
}
