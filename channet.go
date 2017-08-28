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

func (c *client) onOpen(data string)    {}
func (c *client) onClose(data string)   {}
func (c *client) onMessage(data string) {}
func (c *client) onError(data string)   {}

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {

	c, err := s.u.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	go func() {
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
	}()

	go func() {
		for pattern, handler := range handlers {
			for i, stringc := range handler.stringcs {
				err := c.WriteMessage(websocket.TextMessage, []byte(pattern+"$"+strconv.Itoa(i)+"$"+<-stringc))
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	// for {
	// 	msgtype, message, err := c.ReadMessage()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	log.Printf("recv: %s", message)

	// 	err = c.WriteMessage(msgtype, message)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

}

// type safeheap struct {
// 	sync.Mutex
// 	heap []interface{}
// }

// var (
// 	endpoint interface{}

// 	interfaces safeheap
// 	ints       safeheap
// )

// // Interface creates an interface channel in the net space.
// func Interface(buf ...int) (ic chan interface{}) {
// 	interfaces.Lock()
// 	id := len(interfaces.heap)

// 	if len(buf) < 1 {
// 		netsync(netint, 0, id)
// 		ic = make(chan interface{})
// 		interfaces.heap = append(interfaces.heap, nil)
// 	} else {
// 		ic = make(chan interface{}, buf[0])
// 	}

// 	return
// }

// // Int creates an int channel in the net space.
// func Int(buf ...int) (ic chan int) {
// 	ints.Lock()
// 	id := len(ints.heap)

// 	if len(buf) < 1 {
// 		netsync(netint, 0, id)
// 		ic = make(chan int)
// 		ints.heap = append(ints.heap, nil)
// 	} else {
// 		ic = make(chan int, buf[0])
// 	}

// 	return
// }

// type nettype int

// const (
// 	netint nettype = iota
// 	netstring
// )

// func netsync(nt nettype, buf, id int) {
// 	switch nt {
// 	case netint:
// 		switch endpoint.(type) {
// 		case *server:
// 			//
// 			// syncs channel with all client versions
// 			//
// 		case *client:
// 			//
// 			// syncs channel with server's version
// 			//
// 		default:
// 			panic("sync: unable to convert endpoint to type")
// 		}
// 	}
// }
