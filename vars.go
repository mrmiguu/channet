package channet

import "sync"

var (
	handlers = map[string]*Handler{}
	handlerm sync.RWMutex
	handlerc = make(chan *Handler)
)
