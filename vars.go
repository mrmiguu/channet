package channet

import "sync"

var (
	rhandlers = map[string]*Handler{}
	rhandlerm sync.RWMutex
	whandlers = make(chan *Handler)
	sockets   []socket
	socketm   sync.RWMutex
	reboot    = make(chan bool)
)
