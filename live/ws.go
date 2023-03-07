package live

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type wsConn struct {
	sync.Mutex
	*websocket.Conn
}

type message struct {
	Kind    string `json:"kind"`
	Payload string `json:"payload"`
}

func newWsConn(conn *websocket.Conn) *wsConn {
	return &wsConn{sync.Mutex{}, conn}
}

func (ws *wsConn) read() (message, error) {
	ws.Lock()
	defer ws.Unlock()

	var m message
	err := ws.ReadJSON(&m)
	return m, err
}

func (ws *wsConn) write(m any) {
	ws.Lock()
	defer ws.Unlock()
	if err := ws.WriteJSON(m); err != nil {
		log.Println(err)
	}
}
