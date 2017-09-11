package channet

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

// Connect connects one endpoint to another.
func Connect(url string) {
	if strings.Index(url, ":") > 0 {
		ws := js.Global.Get("WebSocket").New("ws://" + url + "/channet")

		ws.Set("onopen", func(evt *js.Object) {
			socketm.Lock()
			sockets = append(sockets, client{ws})
			socketm.Unlock()
		})

		ws.Set("onmessage", func(evt *js.Object) { go read(evt.Get("data").String()) })
	} else {
		s := &server{websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}}
		http.HandleFunc("/channet", s.onConnection)

		go func() {
			err := http.ListenAndServe(url, nil)
			if err != nil {
				panic(err)
			}
		}()
	}

	go write()
}

// New constructs a new base for building network channels.
func New(pattern string) *Handler {
	rhandlerm.Lock()
	defer rhandlerm.Unlock()

	_, exists := rhandlers[pattern]
	if exists {
		panic("handler `" + pattern + "` already exists")
	}

	h := &Handler{
		pattern:  pattern,
		wstrings: make(chan wstring, 1),
		wints:    make(chan wint, 1),
	}
	rhandlers[pattern] = h
	whandlers <- h

	return h
}

// String creates a reading and writing channel for sending strings.
func (h *Handler) String(buf ...int) (<-chan string, chan<- string) {
	if len(buf) > 1 {
		panic("too many arguments")
	}

	l := 0
	if len(buf) > 0 {
		l = buf[0]
	}

	h.rstringm.Lock()
	defer h.rstringm.Unlock()

	r := make(chan string, l)
	w := make(chan string, l)
	i := strconv.Itoa(len(h.rstrings))
	h.rstrings = append(h.rstrings, r)
	h.wstrings <- wstring{i, w}

	return r, w
}

// Int creates a reading and writing channel for sending integers.
func (h *Handler) Int(buf ...int) (<-chan int, chan<- int) {
	if len(buf) > 1 {
		panic("too many arguments")
	}

	l := 0
	if len(buf) > 0 {
		l = buf[0]
	}

	h.rintm.Lock()
	defer h.rintm.Unlock()

	r := make(chan int, l)
	w := make(chan int, l)
	i := strconv.Itoa(len(h.rints))
	h.rints = append(h.rints, r)
	h.wints <- wint{i, w}

	return r, w
}
