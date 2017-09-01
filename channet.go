package channet

import (
	"strings"
)

func Connect(url string) {
	if strings.Index(url, ":") > 0 {
		go initClient(url)
	} else {
		go initServer(url)
	}
}

func New(pattern string) *Handler {
	handlerm.RLock()
	_, exists := handlers[pattern]
	handlerm.RUnlock()

	if exists {
		panic("handler `" + pattern + "` already exists")
	}

	h := &Handler{pattern: pattern, stringcc: make(chan stringc)}

	handlerm.Lock()
	handlers[pattern] = h
	handlerm.Unlock()

	handlerc <- h

	return h
}

func (h *Handler) String(length ...int) (<-chan string, chan<- string) {
	l := 0
	if len(length) > 0 {
		l = length[0]
	}

	r := make(chan string, l)
	w := make(chan string, l)

	h.stringm.Lock()
	i := len(h.strings)
	h.strings = append(h.strings, stringc{i, r})
	h.stringcc <- stringc{i, w}
	h.stringm.Unlock()

	return r, w
}
