package netc

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

type safeheap struct {
	sync.Mutex
	heap []interface{}
}

var (
	endpoint interface{}

	interfaces safeheap
	ints       safeheap
)

// Interface creates an interface channel in the net space.
func Interface(buf ...int) (ic chan interface{}) {
	interfaces.Lock()
	id := len(interfaces.heap)

	if len(buf) < 1 {
		netsync(netint, 0, id)
		ic = make(chan interface{})
		interfaces.heap = append(interfaces.heap, nil)
	} else {
		ic = make(chan interface{}, buf[0])
	}

	return
}

// Int creates an int channel in the net space.
func Int(buf ...int) (ic chan int) {
	ints.Lock()
	id := len(ints.heap)

	if len(buf) < 1 {
		netsync(netint, 0, id)
		ic = make(chan int)
		ints.heap = append(ints.heap, nil)
	} else {
		ic = make(chan int, buf[0])
	}

	return
}

type client struct{}
type server struct {
	u    websocket.Upgrader
	errs []error
}

// Connect creates a network-based channel.
func Connect(addr string) {
	hostPort := strings.Split(addr, ":")
	if len(hostPort) != 2 {
		panic(`Connect addr format must be "[host]:port"`)
	}
	host := hostPort[0]
	port, err := strconv.Atoi(hostPort[1])
	if err != nil {
		panic(err)
	}
	if len(host) < 1 {
		endpoint = initServer(port)
	} else {
		endpoint = initClient(host, port)
	}
}

func initClient(host string, port int) *client {
	sck := js.Global.Get("WebSocket").New("ws://" + host + ":" + strconv.Itoa(port) + "/walk")
	c := &client{}
	sck.Set("onopen", c.onOpen)
	sck.Set("onclose", c.onClose)
	sck.Set("onmessage", c.onMessage)
	sck.Set("onerror", c.onError)
	return c
}

type nettype int

const (
	netint nettype = iota
	netstring
)

func netsync(nt nettype, buf, id int) {
	switch nt {
	case netint:
		switch endpoint.(type) {
		case *server:
			//
			// syncs channel with all client versions
			//
		case *client:
			//
			// syncs channel with server's version
			//
		default:
			panic("sync: unable to convert endpoint to type")
		}
	}
}

func (c *client) onOpen()    {}
func (c *client) onClose()   {}
func (c *client) onMessage() {}
func (c *client) onError()   {}

func initServer(port int) *server {
	s := &server{u: websocket.Upgrader{}}
	http.HandleFunc("/walk", s.onConnection)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *server) onConnection(w http.ResponseWriter, r *http.Request) {
	c, err := s.u.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	for {
		msgtype, message, err := c.ReadMessage()
		if err != nil {
			panic(err)
		}

		log.Printf("recv: %s", message)

		err = c.WriteMessage(msgtype, message)
		if err != nil {
			panic(err)
		}
	}
}
