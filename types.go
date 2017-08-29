package channet

import (
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gorilla/websocket"
)

type client struct {
	ws *js.Object
}

type server struct {
	u    websocket.Upgrader
	errs []error
}

type Handler struct {
	rstrings []rstring
	rstringm sync.RWMutex
	wstrings []wstring
	wstringm sync.RWMutex
}

type rstring struct {
	c   <-chan string
	ref int
}

type wstring struct {
	c   chan<- string
	ref int
}
