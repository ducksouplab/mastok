package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ducksouplab/mastok/env"
	"github.com/ducksouplab/mastok/helpers"
	"github.com/gorilla/websocket"
)

func getHeaderForDialer() http.Header {
	header := http.Header{}
	header.Add("Authorization", "Basic "+helpers.BasicAuth("mastok", "mastok"))
	header.Add("Origin", env.Origin)
	return header
}

func TestCampaignsWS(t *testing.T) {
	t.Run("websocket gives info about an existing campaign", func(t *testing.T) {
		server := httptest.NewServer(getTestRouter())
		defer server.Close()

		// check test_helpers/data.go
		namespace := "namespace"
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/campaigns/supervise?namespace=" + namespace

		ws, res, err := websocket.DefaultDialer.Dial(wsURL, getHeaderForDialer())
		if err != nil {
			if err == websocket.ErrBadHandshake {
				t.Logf("handshake failed with status %d", res.StatusCode)
			}
			t.Fatalf("could not open a ws connection on %s %v", wsURL, err)
		}
		defer ws.Close()

		if err := ws.WriteMessage(websocket.TextMessage, []byte("coucou")); err != nil {
			t.Fatalf("could not send message over ws connection %v", err)
		}

		_, msg, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		t.Log(msg)
	})
}
