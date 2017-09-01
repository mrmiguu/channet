package channet

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
	"github.com/mrmiguu/jsutil"
)

func initClient(url string) *client {
	ws := js.Global.Get("WebSocket").New("ws://" + url + "/channet")

	c := &client{Object: ws}

	ws.Set("onopen", func(evt *js.Object) {
		go func() {
			for {
				write(c)
			}
		}()
	})
	ws.Set("onmessage", func(evt *js.Object) { go read(evt.Get("data").String()) })

	return c
}

func initServer(url string) {
	s := &server{websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}}

	http.HandleFunc("/channet", s.onConnection)

	err := http.ListenAndServe(url, nil)
	if err != nil {
		panic(err)
	}
}

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := connection{conn}

	go func() {
		for {
			_, b, err := c.ReadMessage()
			if err != nil {
				return
			}
			read(string(b))
		}
	}()

	for {
		write(c)
	}
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

func read(packet string) {
	parts := strings.Split(packet, "$")
	pattern, index, message := parts[0], parts[1], parts[2]
	i, err := strconv.Atoi(index)
	if err != nil {
		panic(err)
	}

	handlerm.RLock()
	handlers[pattern].stringm.RLock()
	handlers[pattern].strings[i].c <- message
	handlers[pattern].stringm.RUnlock()
	handlerm.RUnlock()
}

func write(sck socket) {
	for h := range handlerc {
		h := h
		go func() {
			for {
				for sc := range h.stringcc {
					sc := sc
					go func() {
						for {
							err := sck.To(h.pattern + "$" + strconv.Itoa(sc.i) + "$" + <-sc.c)
							if err != nil {
								return
							}
						}
					}()
				}
			}
		}()
	}
}
