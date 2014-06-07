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

func parseMessage(line string) *Message {
	msg := new(Message)
	msg.Raw = line

	// Remove "\r\n"
	if strings.HasSuffix(line, "\r\n") {
		line = line[:len(line)-2]
	}

	// Borrowed from http://git.io/zuwpfA
	if line[0] == ':' {
		if i := strings.Index(line, " "); i > -1 {
			msg.Prefix = line[1:i]
			line = line[i+1:]
		} else {
			log.Printf("Malformed message from server: %#s\n", line)
		}
	}
	split := strings.SplitN(line, " :", 2)
	args := strings.Split(split[0], " ")
	msg.Command = strings.ToUpper(args[0])
	msg.Params = args[1:]
	if len(split) > 1 {
		msg.Params = append(msg.Params, split[1])
	}

	return msg
}
