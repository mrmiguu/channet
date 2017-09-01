package channet

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
	"github.com/mrmiguu/jsutil"
)

func initClient(url string) *client {
	ws := js.Global.Get("WebSocket").New("ws://" + url + "/channet")

	c := &client{Object: ws, msgs: make(chan string, 1)}

	ws.Set("onopen", func(evt *js.Object) { go c.onOpen() })
	ws.Set("onclose", func(evt *js.Object) { go c.onClose() })
	ws.Set("onmessage", func(evt *js.Object) { go c.onMessage(evt) })
	ws.Set("onerror", func(evt *js.Object) { go c.onError(evt) })

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

func (c *client) onOpen() {
	for {
		write(c)
	}
}

func (c *client) onMessage(evt *js.Object) {
	c.msgs <- evt.Get("data").String()
	read(c)
}

func (c *client) onClose()               {}
func (c *client) onError(evt *js.Object) {}

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := connection{conn}

	go func() {
		for {
			read(c)
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

func (c client) From() (string, error) {
	return <-c.msgs, nil
}

func (c client) Print(s string) {
	jsutil.Alert(s)
}

func (c connection) To(packet string) error {
	return c.WriteMessage(websocket.TextMessage, []byte(packet))
}

func (c connection) From() (string, error) {
	_, b, err := c.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c connection) Print(s string) {
	fmt.Println(s)
}

func read(sck socket) {
	pkt, err := sck.From()
	if err != nil {
		return
	}

	parts := strings.Split(pkt, "$")
	pattern, index, message := parts[0], parts[1], parts[2]
	i, err := strconv.Atoi(index)
	if err != nil {
		panic(err)
	}

	handlerm.RLock()
	handlers[pattern].rstringm.RLock()
	handlers[pattern].rstrings[i] <- message
	handlers[pattern].rstringm.RUnlock()
	handlerm.RUnlock()
}

func write(sck socket) {
	handlerm.RLock()
	for pattern, handler := range handlers {
		handler.wstringm.RLock()
		var wg sync.WaitGroup
		wg.Add(len(handler.wstrings))
		for i, wstring := range handler.wstrings {
			i, wstring := i, wstring
			go func() {
				defer wg.Done()
				err := sck.To(pattern + "$" + strconv.Itoa(i) + "$" + <-wstring)
				if err != nil {
					handler.wstringm.RUnlock()
					handlerm.RUnlock()
					return
				}
			}()
		}
		wg.Wait()
		handler.wstringm.RUnlock()
	}
	handlerm.RUnlock()
}
