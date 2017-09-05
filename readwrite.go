package channet

import (
	"strconv"
	"strings"
)

func read(packet string) {
	parts := strings.Split(packet, "$")
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
							err := sck.To(h.pattern + "$" + w.i + "$" + msg)
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
