package main

import (
	"github.com/mrmiguu/channet"
)

func main() {
	channet.Connect(":6969")
	listen()
}

type player struct {
	h *channet.Handler
	r <-chan string
	w chan<- string
}

func listen() {
	db := map[string]player{}

	username, _ := channet.New("login").String()

	for {
		usr := <-username
		msg := "Welcome back, " + usr

		p, exists := db[usr]
		if !exists {
			h := channet.New(usr)
			r, w := h.String()
			p = player{h, r, w}
			db[usr] = p
			msg = "Welcome, " + usr
		}

		p.w <- msg
	}
}
