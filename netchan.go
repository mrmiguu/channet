package netchan

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"test/netchan"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

func example() {

	//
	//
	//

	s := netchan.New(8080)
	<-s.Int() // same channel as client#1 & client#2; blocks until client#1 reads

	//

	c1 := netchan.New(8080, "localhost")
	c1.Int() <- 420 // same channel as client#2 & server; if first, server reads; otherwise, blocks

	//

	c2 := netchan.New(8080, "localhost")
	c2.Int() <- 69 // same channel as client#1 & server; if first, server reads; otherwise, blocks

	//
	//
	//
}

type Type struct {
	endpoint interface{}

	ints     []interface{}
	intsLock sync.Mutex
}

func (t *Type) Int(buf ...int) (ic chan int) {
	t.intsLock.Lock()
	id := len(t.ints)

	if len(buf) < 1 {
		t.sync(netint, 0, id)
		ic = make(chan int)
		t.ints = append(t.ints, nil)
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

// New creates a network-based channel.
func New(port int, host ...string) *Type {
	t := &Type{}
	if len(host) < 1 {
		t.endpoint = initServer(port)
	} else {
		t.endpoint = initClient(host[0], port)
	}
	return t
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

func (t *Type) sync(nt nettype, buf, id int) {
	switch nt {
	case netint:
		switch sc := t.endpoint.(type) {
		case *server:
			//
			// syncs channel with all client versions
			//
		case *client:
			//
			// syncs channel with server's version
			//
		default:
			panic(errors.New("Type.sync: unable to convert endpoint to type"))
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
