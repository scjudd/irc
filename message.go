package main

import (
	"log"
	"strings"
)

type Message struct {
	Raw     string
	Prefix  string
	Command string
	Params  []string
}

func (msg *Message) String() string {
	return msg.Raw
}

func parseMessage(raw string) *Message {
	msg := new(Message)
	msg.Raw = raw

	// Remove "\r\n"
	if strings.HasSuffix(raw, "\r\n") {
		raw = raw[:len(raw)-2]
	}

	// Borrowed from http://git.io/zuwpfA
	if raw[0] == ':' {
		if i := strings.Index(raw, " "); i > -1 {
			msg.Prefix = raw[1:i]
			raw = raw[i+1:]
		} else {
			log.Printf("Malformed message from server: %#s\n", raw)
		}
	}
	split := strings.SplitN(raw, " :", 2)
	args := strings.Split(split[0], " ")
	msg.Command = strings.ToUpper(args[0])
	msg.Params = args[1:]
	if len(split) > 1 {
		msg.Params = append(msg.Params, split[1])
	}

	return msg
}
