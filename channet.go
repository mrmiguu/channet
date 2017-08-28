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
	"sync"

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
	rstrings []<-chan string
	rstringm sync.RWMutex
	wstrings []chan<- string
	wstringm sync.RWMutex
}

func (h *Handler) String() (<-chan string, chan<- string) {
	c := make(chan string)

	h.rstringm.Lock()
	h.rstrings = append(h.rstrings, c)
	h.rstringm.Unlock()

	h.wstringm.Lock()
	h.wstrings = append(h.wstrings, c)
	h.wstringm.Unlock()

	return c, c
}

func New(pattern string) *Handler {
	h := &Handler{}
	handlerm.Lock()
	handlers[pattern] = h
	handlerm.Unlock()
	return h
}

var (
	endpoint interface{}
	handlers map[string]*Handler
	handlerm sync.RWMutex
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
		handlerm.RLock()
		for pattern, handler := range handlers {
			handler.rstringm.RLock()
			for i, rstring := range handler.rstrings {
				c.ws.Call("send", pattern+"$"+strconv.Itoa(i)+"$"+<-rstring)
			}
			handler.rstringm.RUnlock()
		}
		handlerm.RUnlock()
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

	handlerm.RLock()
	handlers[pattern].wstringm.RLock()
	handlers[pattern].wstrings[i] <- message
	handlers[pattern].wstringm.RUnlock()
	handlerm.RUnlock()

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

			handlerm.RLock()
			handlers[pattern].wstringm.RLock()
			handlers[pattern].wstrings[i] <- message
			handlers[pattern].wstringm.RUnlock()
			handlerm.RUnlock()
		}
	}()

	go func() {
		for {
			handlerm.RLock()
			for pattern, handler := range handlers {
				handler.rstringm.RLock()
				for i, rstring := range handler.rstrings {
					err := c.WriteMessage(websocket.TextMessage, []byte(pattern+"$"+strconv.Itoa(i)+"$"+<-rstring))
					if err != nil {
						panic(err)
					}
				}
				handler.rstringm.RUnlock()
			}
			handlerm.RUnlock()
		}
	}()

}
