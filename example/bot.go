package main

import (
	"log"
	"strings"

	"github.com/scjudd/irc"
)

func main() {
	irc.EnableLogging()
	c := irc.NewConnection()

	// Join #bot once handshake completes
	c.RegisterHandler("001", func(msg *irc.Message) bool {
		c.Join("#bot")
		return true
	})

	// Join an arbitrary channel on invitation
	c.RegisterHandler("INVITE", func(msg *irc.Message) bool {
		c.Join(msg.Params[1])
		return true
	})

	c.RegisterHandler("PRIVMSG", func(msg *irc.Message) bool {
		if strings.HasPrefix(msg.Params[1], "!") {
			// Send raw command to server
			c.SendRawMessage(msg.Params[1][1:] + "\r\n")
		} else {
			// Echo back every message
			c.Privmsg(msg.Params[0], msg.Params[1])
		}
		return true
	})

	if err := c.Connect("localhost:6667", "bot"); err != nil {
		log.Fatal(err)
	}
}
