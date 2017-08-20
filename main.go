package main

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/mrmiguu/netc/netc"
	"github.com/mrmiguu/netc/safedb"
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

//
//
//

// there is one of this

func testServer() {
	netc.Connect(":6969")

	db := safedb.New()

	go func() {
		ic := netc.Interface()
		lgn := (<-ic).(login)

		db.Put(lgn.key, ic)

		for {
			msg := (<-ic).(message)

			if len(msg.to) < 1 {
				fmt.Println(msg.body) // no 'to'; meant for server
				continue
			}

			c, found := db.Lookup(msg.to)

			if !found {
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

	ic <- message{"alxndrthegrt91", "hey there, bud!"}

	js.Global.Call("alert", <-ic)
}

//
//
//
