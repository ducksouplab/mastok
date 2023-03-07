package types

type Message struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`
}
