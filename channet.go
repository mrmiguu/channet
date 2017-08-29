package channet

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
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

	// js.Global.Call("alert", "New :: handlerm.Lock()...")
	// fmt.Println(`New handler lock...`)
	handlerm.Lock()
	// fmt.Println(`New handler lock !`)
	// js.Global.Call("alert", "New :: handlerm.Lock() !")
	handlers[pattern] = h
	// fmt.Println(`handlers[`+pattern+`] =`, handlers[pattern])
	handlerm.Unlock()

	return h
}

func (h *Handler) String(length ...int) (<-chan string, chan<- string) {

	l := 0
	if len(length) > 0 {
		l = length[0]
	}

	rw := make(chan string, l)

	// js.Global.Call("alert", "String :: h.rstringm.Lock()...")
	h.rstringm.Lock()
	// js.Global.Call("alert", "String :: h.rstringm.Lock()!")
	h.rstrings = append(h.rstrings, rstring{rw, 1})
	h.rstringm.Unlock()

	// js.Global.Call("alert", "String :: h.wstringm.Lock()...")
	h.wstringm.Lock()
	// js.Global.Call("alert", "String :: h.wstringm.Lock()!")
	h.wstrings = append(h.wstrings, wstring{rw, 1})
	h.wstringm.Unlock()

	return rw, rw
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

func initServer(url string) *server {

	s := &server{u: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}}

	http.HandleFunc("/channet", s.onConnection)

	err := http.ListenAndServe(url, nil)
	if err != nil {
		panic(err)
	}

	return s
}

func (c *client) onOpen(data string) {

	for {
		// js.Global.Call("alert", "onOpen :: handlerm.Lock()...")
		handlerm.RLock()
		// js.Global.Call("alert", "len(handlers)="+strconv.Itoa(len(handlers)))
		// js.Global.Call("alert", "onOpen :: handlerm.Lock() !")
		for pattern, handler := range handlers {
			// js.Global.Call("alert", "onOpen :: handler.rstringm.Lock()...")
			handler.rstringm.RLock()
			// js.Global.Call("alert", "onOpen :: handler.rstringm.Lock() !")
			for i, rstring := range handler.rstrings {
				// js.Global.Call("alert", "onOpen :: c.ws.Call(\"send\",...)...")
				c.ws.Call("send", pattern+"$"+strconv.Itoa(i)+"$"+<-rstring.c)
				// js.Global.Call("alert", "onOpen :: c.ws.Call(\"send\",...) !")
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

	// js.Global.Call("alert", "onMessage :: handlerm.Lock()...")
	handlerm.RLock()
	// js.Global.Call("alert", "onMessage :: handlerm.Lock() !")
	// js.Global.Call("alert", "onMessage :: handlers[pattern].wstringm.Lock()...")
	handlers[pattern].wstringm.RLock()
	// js.Global.Call("alert", "onMessage :: handlers[pattern].wstringm.Lock() !")
	handlers[pattern].wstrings[i].c <- message
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
			// fmt.Println(`c.ReadMessage()...`)
			_, b, err := c.ReadMessage()
			if err != nil {
				return
			}
			// fmt.Println(`c.ReadMessage() !`, string(b))

			parts := strings.Split(string(b), "$")
			pattern, index, message := parts[0], parts[1], parts[2]
			i, err := strconv.Atoi(index)
			if err != nil {
				panic(err)
			}

			// fmt.Println("pattern ::", pattern)

			handlerm.RLock()
			handlers[pattern].wstringm.RLock()
			handlers[pattern].wstrings[i].c <- message
			handlers[pattern].wstringm.RUnlock()
			handlerm.RUnlock()
		}
	}()

	for {
		handlerm.RLock()
		for pattern, handler := range handlers {
			handler.rstringm.RLock()
			for i, rstring := range handler.rstrings {
				// fmt.Println(`c.WriteMessage()...`)
				err := c.WriteMessage(websocket.TextMessage, []byte(pattern+"$"+strconv.Itoa(i)+"$"+<-rstring.c))
				if err != nil {
					return
				}
				// fmt.Println(`c.WriteMessage() !`)
			}
			handler.rstringm.RUnlock()
		}
		handlerm.RUnlock()
	}
}
