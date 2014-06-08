package irc

import (
	"log"
)

type MessageDispatcher interface {
	RegisterHandler(string, func(*Message) bool)
	Dispatch(*Message)
}

type SimpleDispatcher map[string][]func(*Message) bool

func NewDispatcher() SimpleDispatcher {
	return make(SimpleDispatcher)
}

func (sd SimpleDispatcher) RegisterHandler(cmd string, handler func(*Message) bool) {
	sd[cmd] = append(sd[cmd], handler)
}

func (sd SimpleDispatcher) Dispatch(msg *Message) {
	handled := false
	for _, handler := range sd[msg.Command] {
		if handler(msg) {
			handled = true
		}
	}

	if handled {
		log.Printf("\x1b[96;40;1m<-- %s\x1b[0m\n", msg.Raw[:len(msg.Raw)-2])
	} else {
		log.Printf("<-- %s\n", msg.Raw[:len(msg.Raw)-2])
	}
}
