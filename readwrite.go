package channet

import (
	"strconv"
	"strings"
)

const (
	us31 = "▼"
)

func read(packet string) {
	parts := strings.Split(packet, us31)
	pattern, idx, msg := parts[0], parts[1], parts[2]
	i, err := strconv.Atoi(idx)
	if err != nil {
		panic(err)
	}

	rhandlerm.RLock()
	defer rhandlerm.RUnlock()

	_, exists := rhandlers[pattern]
	if !exists {
		return
	}

	h := rhandlers[pattern]

	h.rstringm.RLock()
	defer h.rstringm.RUnlock()

	if i >= len(h.rstrings) {
		return
	}

	h.rstrings[i] <- msg
}

func write() {
	for h := range whandlers {
		go func(h *Handler) {
			for w := range h.wstrings {
				go func(w wstring) {
					for msg := range w.c {
						socketm.RLock()
						for _, sck := range sockets {
							err := sck.To(h.pattern + us31 + w.i + us31 + msg)
							if err != nil {
								continue
							}
						}
						socketm.RUnlock()
					}
				}(w)
			}
		}(h)
	}
}
