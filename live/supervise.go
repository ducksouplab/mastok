package live

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type supervisor struct {
	sync.Mutex
	*websocket.Conn
}

func newSupervisor(conn *websocket.Conn) *supervisor {
	return &supervisor{sync.Mutex{}, conn}
}

func (s *supervisor) loop() {

}

func Supervise(conn *websocket.Conn, namespace string) {
	log.Println("[supervise] running for: " + namespace)
	// TODO load model
	s := newSupervisor(conn)
	s.loop() //
}
