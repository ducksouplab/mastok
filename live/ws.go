package live

import (
	"log"
	"sync"

	"github.com/ducksouplab/mastok/types"
	"github.com/gorilla/websocket"
)

type wsConn struct {
	sync.Mutex
	*websocket.Conn
}

func newWsConn(conn *websocket.Conn) *wsConn {
	return &wsConn{sync.Mutex{}, conn}
}

func (ws *wsConn) read() (types.Message, error) {
	// don't lock on read since ReadJSON will wait until a message arrives
	var m types.Message
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
