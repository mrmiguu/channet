package channet

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	us31 = "▼"
)

func read(packet string) {
	parts := strings.Split(packet, us31)
	pattern, idx, mtype, msg := parts[0], parts[1], parts[2], parts[3]
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

	switch mtype {
	case tbool:
	case tstring:
		h.rstringm.RLock()
		defer h.rstringm.RUnlock()
		if i >= len(h.rstrings) {
			return
		}
		h.rstrings[i] <- msg
	case tint:
		h.rintm.RLock()
		defer h.rintm.RUnlock()
		if i >= len(h.rints) {
			return
		}
		x, err := strconv.Atoi(msg)
		if err != nil {
			panic("cannot convert message to int")
		}
		h.rints[i] <- x
	case tint8:
	case tint16:
	case tint32:
	case tint64:
	case tuint:
	case tuint8:
	case tuint16:
	case tuint32:
	case tuint64:
	case tbyte:
	case trune:
	case tfloat32:
	case tfloat64:
	case tcomplex64:
	case tcomplex128:
	default:
		panic("unknown type")
	}
}

func write() {
	for h := range whandlers {
		go func(h *Handler) {
			for w := range h.wstrings {
				go func(w wstring) {
					for msg := range w.c {
						socketm.RLock()
						if len(sockets) < 1 {
							socketm.RUnlock()
							<-reboot
							socketm.RLock()
						}
						for _, sck := range sockets {
							err := sck.To(h.pattern + us31 + w.i + us31 + tstring + us31 + msg)
							if err != nil {
								continue
							}
						}
						socketm.RUnlock()
					}
				}(w)
			}
		}(h)
		go func(h *Handler) {
			for w := range h.wints {
				go func(w wint) {
					for msg := range w.c {
						socketm.RLock()
						if len(sockets) < 1 {
							socketm.RUnlock()
							<-reboot
							fmt.Println("REBOOT")
							socketm.RLock()
						}
						dead := []int{}
						for i, sck := range sockets {
							err := sck.To(h.pattern + us31 + w.i + us31 + tint + us31 + strconv.Itoa(msg))
							if err != nil {
								panic("error writing")
								dead = append(dead, i)
							}
						}
						fmt.Println("dead=", len(dead))
						for _, i := range dead {
							copy(sockets[i:], sockets[i+1:])
							sockets[len(sockets)-1] = nil
							sockets = sockets[:len(sockets)-1]
						}
						socketm.RUnlock()
					}
				}(w)
			}
		}(h)
	}
}
