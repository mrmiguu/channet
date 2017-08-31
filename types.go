package channet

import (
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

type Handler struct {
	rstrings []chan string
	rstringm sync.RWMutex
	wstrings []chan string
	wstringm sync.RWMutex
}

type client struct {
	*js.Object
	msgs chan string
}

type server struct {
	websocket.Upgrader
}

type connection struct {
	*websocket.Conn
}

type socket interface {
	To(string) error
	From() (string, error)
}
