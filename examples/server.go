package main

import (
	"github.com/mrmiguu/channet"
)

func main() {
	channet.Connect(":6969")
	listen()
}

type user struct {
	h *channet.Handler
	r <-chan string
	w chan<- string
}

func listen() {
	db := map[string]user{}

	lr, _ := channet.New("login").String()

	for {
		usr := <-lr
		msg := "Welcome back, " + usr

		u, exists := db[usr]
		if !exists {
			uh := channet.New(usr)
			ur, uw := uh.String()
			u = user{uh, ur, uw}
			db[usr] = u
			msg = "Welcome, " + usr
		}

		u.w <- msg
	}
}
