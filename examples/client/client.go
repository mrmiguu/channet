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
	_, lw := channet.New("login").String()
	usr := jsutil.Prompt("username?")
	lw <- usr
	ur, _ := channet.New(usr).String()
	jsutil.Alert(<-ur)
}
