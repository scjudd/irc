package main

import (
	"log"
	"strings"
)

func main() {
	c := NewConnection()

	// Join #bot once handshake completes
	c.RegisterHandler("001", func(msg *Message) {
		c.Join("#bot")
	})

	// Join an arbitrary channel on invitation
	c.RegisterHandler("INVITE", func(msg *Message) {
		c.Join(msg.Params[1])
	})

	c.RegisterHandler("PRIVMSG", func(msg *Message) {
		if strings.HasPrefix(msg.Params[1], "!") {
			// Send raw command to server
			c.SendString(msg.Params[1][1:] + "\r\n")
		} else {
			// Echo back every message
			c.Privmsg(msg.Params[0], msg.Params[1])
		}
	})

	if err := c.Connect("localhost:6667", "bot"); err != nil {
		log.Fatal(err)
	}
}
