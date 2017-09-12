package channet

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mrmiguu/jsutil"
)

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	socketm.Lock()
	i := len(sockets)
	fmt.Println("socket[", i, "] !")
	if i < 1 {
		fmt.Println("reboot <- true...")
		reboot <- true
		fmt.Println("reboot <- true !")
	}
	sockets = append(sockets, connection{conn})
	socketm.Unlock()

	go func() {
		for {
			_, b, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("conn.ReadMessage() error")
				conn.Close()

				socketm.Lock()
				copy(sockets[i:], sockets[i+1:])
				sockets[len(sockets)-1] = nil
				sockets = sockets[:len(sockets)-1]
				socketm.Unlock()

				return
			}
			read(string(b))
		}
	}()
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
