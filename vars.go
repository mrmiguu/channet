package channet

import "sync"

var (
	rhandlers = map[string]*Handler{}
	rhandlerm sync.RWMutex
	whandlers = make(chan *Handler, 1)
	sockets   []socket
	socketm   sync.RWMutex
)
