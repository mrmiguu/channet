package channet

import (
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

// Handler is a base for constructing network channels.
type Handler struct {
	pattern  string
	rstrings []chan string
	rstringm sync.RWMutex
	wstrings []chan string
	wstringm sync.RWMutex
}

type client struct {
	*js.Object
}

type server struct {
	websocket.Upgrader
}

type connection struct {
	*websocket.Conn
}

type socket interface {
	To(string) error
	Print(string)
}
