package irc

import (
	"log"
)

type MessageDispatcher interface {
	RegisterHandler(string, func(*Message))
	Dispatch(*Message)
}

type SimpleDispatcher map[string][]func(*Message)

func NewDispatcher() SimpleDispatcher {
	return make(SimpleDispatcher)
}

func (sd SimpleDispatcher) RegisterHandler(cmd string, handler func(*Message)) {
	sd[cmd] = append(sd[cmd], handler)
}

func (sd SimpleDispatcher) Dispatch(msg *Message) {
	if len(sd[msg.Command]) > 0 {
        log.Printf("\x1b[96;40;1m<-- %s\x1b[0m\n", msg.Raw[:len(msg.Raw)-2])
		for _, handler := range sd[msg.Command] {
			handler(msg)
		}
	} else {
		log.Printf("<-- %s\n", msg.Raw[:len(msg.Raw)-2])
	}
}
