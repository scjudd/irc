package irc

import (
	"log"
	"strings"
)

type Message struct {
	Raw     string
	Prefix  string
	Nick    string
	User    string
	Host    string
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
		if i, j := strings.Index(msg.Prefix, "!"), strings.Index(msg.Prefix, "@"); i > -1 && j > -1 {
			msg.Nick = msg.Prefix[0:i]
			msg.User = msg.Prefix[i+1 : j]
			msg.Host = msg.Prefix[j+1 : len(msg.Prefix)]
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
