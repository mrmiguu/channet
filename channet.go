//
//
//

// IFF: data race due to socket R/W on interfacing channel
// (fix) when reading from channel, select based on gid;
//       the gid of the socket-writing goroutine will wait
//       until the user has read the channel and then written
//       to it, releasing it out onto the wire

//
//
//

package channet

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

func Connect(url string) {

	if strings.Index(url, ":") > 0 {
		endpoint = initClient(url)
	} else {
		endpoint = initServer(url)
	}

}

type Handler struct {
	stringcs []chan string
}

func (h *Handler) String() chan string {
	stringc := make(chan string)
	h.stringcs = append(h.stringcs, stringc)
	return stringc
}

func New(pattern string) *Handler {
	h := Handler{}
	handlers[pattern] = h
	return &h
}

var (
	endpoint interface{}
	handlers map[string]Handler
)

type client struct {
	ws *js.Object
}

func initClient(url string) *client {

	ws := js.Global.Get("WebSocket").New("ws://" + url + "/channet")

	c := &client{ws: ws}

	ws.Set("onopen", func(evt *js.Object) { c.onOpen(evt.Get("data").String()) })
	ws.Set("onclose", func(evt *js.Object) { c.onClose(evt.Get("data").String()) })
	ws.Set("onmessage", func(evt *js.Object) { c.onMessage(evt.Get("data").String()) })
	ws.Set("onerror", func(evt *js.Object) { c.onError(evt.Get("data").String()) })

	return c

}

type server struct {
	u    websocket.Upgrader
	errs []error
}

func initServer(url string) *server {

	s := &server{u: websocket.Upgrader{}}

	http.HandleFunc("/channet", s.onConnection)

	err := http.ListenAndServe(url, nil)
	if err != nil {
		panic(err)
	}

	return s

}

func (c *client) onOpen(data string) {

	for {
		for pattern, handler := range handlers {
			for i, stringc := range handler.stringcs {
				c.ws.Call("send", pattern+"$"+strconv.Itoa(i)+"$"+<-stringc)
			}
		}
	}

}

func (c *client) onClose(data string) {}

func (c *client) onMessage(data string) {

	parts := strings.Split(data, "$")
	pattern, index, message := parts[0], parts[1], parts[2]
	i, err := strconv.Atoi(index)
	if err != nil {
		panic(err)
	}

	handlers[pattern].stringcs[i] <- message

}

func (c *client) onError(data string) {}

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {

	c, err := s.u.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	go func() {
		for {
			_, b, err := c.ReadMessage()
			if err != nil {
				panic(err)
			}

			parts := strings.Split(string(b), "$")
			pattern, index, message := parts[0], parts[1], parts[2]
			i, err := strconv.Atoi(index)
			if err != nil {
				panic(err)
			}

			handlers[pattern].stringcs[i] <- message
		}
	}()

	go func() {
		for {
			for pattern, handler := range handlers {
				for i, stringc := range handler.stringcs {
					err := c.WriteMessage(websocket.TextMessage, []byte(pattern+"$"+strconv.Itoa(i)+"$"+<-stringc))
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}()

}
