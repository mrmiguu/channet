package channet

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mrmiguu/jsutil"
)

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Println("onConnection!")
	conn, err := s.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			_, b, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("read error")
				conn.Close()
				return
			}
			read(string(b))
		}
	}()

	socketm.Lock()
	if len(sockets) < 1 {
		reboot <- true
	}
	sockets = append(sockets, connection{conn})
	socketm.Unlock()
}

func (c client) To(packet string) (err error) {
	defer jsutil.OnPanic(&err)
	c.Call("send", packet)
	return
}

func (c client) Print(s string) {
	jsutil.Alert(s)
}

func (c connection) To(packet string) error {
	return c.WriteMessage(websocket.TextMessage, []byte(packet))
}

func (c connection) Print(s string) {
	fmt.Println(s)
}
