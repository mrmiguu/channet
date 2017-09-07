package main

import (
	"github.com/mrmiguu/channet"
	"github.com/mrmiguu/jsutil"
)

func main() {
	channet.Connect("localhost:6969")
	login()
}

func login() {
	_, username := channet.New("login").String()
	usr := jsutil.Prompt("username?")
	username <- usr
	r, _ := channet.New(usr).String()
	jsutil.Alert(<-r)
}
