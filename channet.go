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

	// if js.Global != nil && js.Global.Call != nil {
	// 	js.Global.Call("alert", "New :: handlerm.RLock()...")
	// }
	handlerm.RLock()
	// if js.Global != nil && js.Global.Call != nil {
	// 	js.Global.Call("alert", "New :: handlerm.RLock() !")
	// }
	_, exists := handlers[pattern]
	handlerm.RUnlock()

	if exists {
		panic("handler `" + pattern + "` already exists")
	}

	h := &Handler{}

	// if js.Global != nil && js.Global.Call != nil {
	// 	js.Global.Call("alert", "New :: handlerm.Lock()...")
	// }
	// fmt.Println(`New handler lock...`)
	handlerm.Lock()
	// if js.Global != nil && js.Global.Call != nil {
	// 	js.Global.Call("alert", "New :: handlerm.Lock() !")
	// }
	// fmt.Println(`New handler lock !`)
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

	r := make(chan string, l)
	w := make(chan string, l)

	// js.Global.Call("alert", "String :: h.rstringm.Lock()...")
	h.rstringm.Lock()
	// js.Global.Call("alert", "String :: h.rstringm.Lock()!")
	h.rstrings = append(h.rstrings, rstring{r, 1})
	h.rstringm.Unlock()

	// js.Global.Call("alert", "String :: h.wstringm.Lock()...")
	h.wstringm.Lock()
	// js.Global.Call("alert", "String :: h.wstringm.Lock()!")
	h.wstrings = append(h.wstrings, wstring{w, 1})
	h.wstringm.Unlock()

	return r, w
}

func initClient(url string) *client {

	ws := js.Global.Get("WebSocket").New("ws://" + url + "/channet")

	c := &client{ws: ws}

	ws.Set("onopen", func(evt *js.Object) { go c.onOpen(evt.Get("data").String()) })
	ws.Set("onclose", func(evt *js.Object) { go c.onClose(evt.Get("data").String()) })
	ws.Set("onmessage", func(evt *js.Object) { go c.onMessage(evt.Get("data").String()) })
	ws.Set("onerror", func(evt *js.Object) { go c.onError(evt.Get("data").String()) })

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
			// js.Global.Call("alert", "onOpen :: handler.wstringm.Lock()...")
			handler.wstringm.RLock()
			// js.Global.Call("alert", "onOpen :: handler.wstringm.Lock() !")
			for i, wstring := range handler.wstrings {
				select {
				case s := <-wstring.c:
					// js.Global.Call("alert", "sending `"+s+"`")
					c.ws.Call("send", pattern+"$"+strconv.Itoa(i)+"$"+s)
					// js.Global.Call("alert", "send !")
					// default:
				}
			}
			handler.wstringm.RUnlock()
		}
		handlerm.RUnlock()
	}
}

func (c *client) onClose(data string) {
	// js.Global.Call("alert", "[WS CLOSED]")
}

func (c *client) onMessage(data string) {

	// js.Global.Call("alert", "onMessage !")
	parts := strings.Split(data, "$")
	pattern, index, message := parts[0], parts[1], parts[2]
	i, err := strconv.Atoi(index)
	if err != nil {
		panic(err)
	}

	// js.Global.Call("alert", "onMessage :: handlerm.Lock()...")
	handlerm.RLock()
	// js.Global.Call("alert", "onMessage :: handlerm.Lock() !")
	// js.Global.Call("alert", "onMessage :: handlers[pattern].rstringm.Lock()...")
	handlers[pattern].rstringm.RLock()
	// js.Global.Call("alert", "onMessage :: handlers[pattern].rstringm.Lock() !")
	// js.Global.Call("alert", "onMessage :: handlers[pattern].rstrings[i].c <- message...")
	handlers[pattern].rstrings[i].c <- message
	// js.Global.Call("alert", "onMessage :: handlers[pattern].rstrings[i].c <- message !")
	handlers[pattern].rstringm.RUnlock()
	handlerm.RUnlock()
}

func (c *client) onError(data string) {
	// js.Global.Call("alert", "[WS ERROR]")
}

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
			handlers[pattern].rstringm.RLock()
			handlers[pattern].rstrings[i].c <- message
			handlers[pattern].rstringm.RUnlock()
			handlerm.RUnlock()
		}
	}()

	for {
		handlerm.RLock()
		for pattern, handler := range handlers {
			handler.wstringm.RLock()
			for i, wstring := range handler.wstrings {
				// fmt.Println(`len(handler.wstrings)=`, len(handler.wstrings))
				select {
				case s := <-wstring.c:
					// fmt.Println("sending `" + s + "`")
					// fmt.Println(`c.WriteMessage()...`)
					err := c.WriteMessage(websocket.TextMessage, []byte(pattern+"$"+strconv.Itoa(i)+"$"+s))
					if err != nil {
						return
					}
					// fmt.Println(`c.WriteMessage() !`)
					// default:
				}
			}
			handler.wstringm.RUnlock()
		}
		handlerm.RUnlock()
	}
}
