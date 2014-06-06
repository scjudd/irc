package main

import (
	"log"
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

	// Echo back every message
	c.RegisterHandler("PRIVMSG", func(msg *Message) {
		c.Privmsg(msg.Params[0], msg.Params[1])
	})

	if err := c.Connect("localhost:6667", "bot"); err != nil {
		log.Fatal(err)
	}
}
