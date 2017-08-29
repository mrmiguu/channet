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
		go initClient(url)
	} else {
		go initServer(url)
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

	// js.Global.Call("alert", "String :: h.rstringm.Lock()...")
	h.rstringm.Lock()
	// js.Global.Call("alert", "String :: h.rstringm.Lock()!")
	h.rstrings = append(h.rstrings, c)
	h.rstringm.Unlock()

	// js.Global.Call("alert", "String :: h.wstringm.Lock()...")
	h.wstringm.Lock()
	// js.Global.Call("alert", "String :: h.wstringm.Lock()!")
	h.wstrings = append(h.wstrings, c)
	h.wstringm.Unlock()

	return c, c
}

func New(pattern string) *Handler {
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

var (
	endpoint interface{}
	handlers = map[string]*Handler{}
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
				c.ws.Call("send", pattern+"$"+strconv.Itoa(i)+"$"+<-rstring)
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
			handlers[pattern].wstrings[i] <- message
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
				err := c.WriteMessage(websocket.TextMessage, []byte(pattern+"$"+strconv.Itoa(i)+"$"+<-rstring))
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
