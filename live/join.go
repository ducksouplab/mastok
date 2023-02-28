package live

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type joiner struct {
	sync.Mutex
	*websocket.Conn
}

func newJoiner(conn *websocket.Conn) *joiner {
	return &joiner{sync.Mutex{}, conn}
}

func (j *joiner) loop() {

}

func Join(conn *websocket.Conn, namespace string) {
	log.Println("[join] running for: " + namespace)
	// TODO load model
	j := newJoiner(conn)
	j.loop() //
}
