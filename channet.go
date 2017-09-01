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

	h := &Handler{}

	handlerm.Lock()
	handlers[pattern] = h
	handlerm.Unlock()

	return h
}

func (h *Handler) String(length ...int) (<-chan string, chan<- string) {
	l := 0
	if len(length) > 0 {
		l = length[0]
	}

	r := make(chan string, l)
	w := make(chan string, l)

	h.rstringm.Lock()
	h.rstrings = append(h.rstrings, r)
	h.rstringm.Unlock()

	h.wstringm.Lock()
	h.wstrings = append(h.wstrings, w)
	h.wstringm.Unlock()

	return r, w
}
