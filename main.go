package main

import (
	"fmt"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/mrmiguu/netc/netc"
)

func main() {
	go testServer()
	go testClient()
	go testClient()
	go testClient()
}

//
//
//

// this is a shared structure

type login struct {
	key    string
	secret string
}

type message struct {
	to   string
	body string
}

type safedb struct {
	sync.RWMutex
	m map[string]chan interface{}
}

// there is one of this

func testServer() {
	netc.Connect(":6969")

	db := safedb{m: make(map[string]chan interface{})}

	go func() {
		ic := netc.Interface()
		lgn := (<-ic).(login)

		db.Lock()
		db.m[lgn.key] = ic
		db.Unlock()

		for {
			msg := (<-ic).(message)

			if len(msg.to) < 1 {
				fmt.Println(msg.body) // no 'to'; meant for server
				continue
			}

			db.RLock()
			c, userFound := db.m[msg.to]
			db.RUnlock()

			if !userFound {
				ic <- "user '" + msg.to + "' was not found"
				continue
			}

			c <- lgn.key + `: "` + msg.body + `"` // response to player
		}
	}()
}

// there are three of these

func testClient() {
	netc.Connect("localhost:6969")
	ic := netc.Interface()
	ic <- login{"alxndrthegrt91", "*******"}
	ic <- message{body: "hey there!"}
	js.Global.Call("alert", <-ic)
}

//
//
//
