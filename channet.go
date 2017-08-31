package channet

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
	"github.com/mrmiguu/jsutil"
)

var (
	handlers = map[string]*Handler{}
	handlerm sync.RWMutex
)

func Connect(url string) {
	if strings.Index(url, ":") > 0 {
		go initClient(url)
	} else {
		go initServer(url)
	}
}

func New(pattern string) *Handler {
	handlerm.RLock()
	_, exists := handlers[pattern]
	handlerm.RUnlock()

	if exists {
		panic("handler `" + pattern + "` already exists")
	}

	h := &Handler{}

	handlerm.Lock()
	handlers[pattern] = h
	handlerm.Unlock()

	return h
}

func (h *Handler) String(length ...int) (<-chan string, chan<- string) {
	l := 0
	if len(length) > 0 {
		l = length[0]
	}

	r := make(chan string, l)
	w := make(chan string, l)

	h.rstringm.Lock()
	h.rstrings = append(h.rstrings, r)
	h.rstringm.Unlock()

	h.wstringm.Lock()
	h.wstrings = append(h.wstrings, w)
	h.wstringm.Unlock()

	return r, w
}

func (c client) To(packet string) (err error) {
	defer jsutil.OnPanic(&err)
	c.Call("send", packet)
	return
}

func (c connection) To(packet string) error {
	return c.WriteMessage(websocket.TextMessage, []byte(packet))
}

func (c client) From() (string, error) {
	return <-c.msgs, nil
}

func (c connection) From() (string, error) {
	_, b, err := c.ReadMessage()
	if err != nil {
		return "", err
	}
	return string(b), nil
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
		for i, wstring := range handler.wstrings {
			select {
			case s := <-wstring:
				sck.To(pattern + "$" + strconv.Itoa(i) + "$" + s)
			default:
			}
		}
		handler.wstringm.RUnlock()
	}
	handlerm.RUnlock()
}

func initClient(url string) *client {
	ws := js.Global.Get("WebSocket").New("ws://" + url + "/channet")

	c := &client{Object: ws, msgs: make(chan string)}

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

func (c *client) onClose() {}

func (c *client) onMessage(evt *js.Object) {
	c.msgs <- evt.Get("data").String()
}

func (c *client) onError(evt *js.Object) {}

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := connection{conn}

	go write(c)

	for {
		read(c)
	}
}
