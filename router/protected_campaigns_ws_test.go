package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/helpers"
	"github.com/ducksouplab/mastok/live"
	"github.com/gorilla/websocket"
)

func dial(t *testing.T, server *httptest.Server, path string) *websocket.Conn {
	url := "ws" + strings.TrimPrefix(server.URL, "http") + path
	header := http.Header{}
	header.Add("Authorization", "Basic "+helpers.BasicAuth("mastok", "mastok"))
	header.Add("Origin", env.Origin)

	ws, res, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			t.Fatalf("ws handshake on %s failed with status %d", url, res.StatusCode)
		}
		t.Fatalf("ws connection on %s failed %v", url, err)
	}
	return ws
}

func TestCampaignsSuperviseWS_Integration(t *testing.T) {
	t.Run("websocket gives info about an existing campaign", func(t *testing.T) {
		server := httptest.NewServer(getTestRouter())
		defer server.Close()

		// check test_helpers/data.go
		namespace := "fxt_router_ns1"

		ws := dial(t, server, "/ws/campaigns/supervise?namespace="+namespace)
		defer ws.Close()

		if err := ws.WriteJSON(live.Message{Kind: "State", Payload: "Running"}); err != nil {
			t.Fatalf("%v", err)
		}

		var reply live.Message
		err := ws.ReadJSON(&reply)
		if err != nil {
			t.Fatalf("%v", err)
		}
		// assert something?
	})
}
