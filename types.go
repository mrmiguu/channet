package channet

import (
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

// Handler is a base for constructing network channels.
type Handler struct {
	pattern string

	rstrings []chan string
	rstringm sync.RWMutex
	wstrings chan wstring

	rints []chan int
	rintm sync.RWMutex
	wints chan wint
}

type wstring struct {
	i string
	c chan string
}
type wint struct {
	i string
	c chan int
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
