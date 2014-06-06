package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type Connection struct {
	sock        net.Conn
	nick        string
	read, write chan string
	MessageDispatcher
}

func NewConnection() *Connection {
	read := make(chan string)
	write := make(chan string)
	dispatcher := NewDispatcher()
	return &Connection{nil, "", read, write, dispatcher}
}

// c.Connect("irc.hashbang.sh:6667", "bot")
func (c *Connection) Connect(server, nick string) error {
	if c.sock != nil {
		return errors.New("Connection already established")
	}

	if c.MessageDispatcher == nil {
		return errors.New("Connection's MessageDispatcher must be instantiated")
	}

	c.RegisterHandler("PING", func(msg *Message) {
		c.WriteString("PONG " + strings.Join(msg.Params, " "))
	})

	sock, err := net.Dial("tcp", server)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	shutdownChan := make(chan struct{})
	fatal := func(msg string) {
		errChan <- errors.New(msg)
		close(shutdownChan)
	}

	// read goroutine
	go func() {
		br := bufio.NewReaderSize(sock, MaxMsgLen)
		for {
			select {
			case <-shutdownChan:
				return
			default:
				str, err := br.ReadString('\n')
				if err != nil {
					// TODO(scjudd): if err is EOF, try reconnecting
					fatal(fmt.Sprintf("read: %s", err))
					return
				}
				str = str[:len(str)-2] // remove trailing "\r\n"
				c.read <- str
			}
		}
	}()

	// write goroutine
	go func() {
		bw := bufio.NewWriterSize(sock, MaxMsgLen)
		for {
			select {
			case <-shutdownChan:
				return
			default:
				str, ok := <-c.write
				if !ok {
					fatal("write: channel closed")
					return
				}
				_, err := bw.WriteString(str + "\r\n")
				if err != nil {
					fatal(fmt.Sprintf("write: %s", err))
					return
				}
				bw.Flush()
				log.Printf("\x1b[92;40;1m--> %s\x1b[0m\n", str)
			}
		}
	}()

	// dispatcher
	go func() {
		// TODO(scjudd): periodically send PINGs
		for {
			select {
			case <-shutdownChan:
				return
			default:
				msg := parseMessage(<-c.read)
				// TODO(scjudd): proper prefix parsing, so someone with nick botbot won't get ignored
				if strings.Index(msg.Prefix, nick) == 0 {
					continue // ignore our own messages
				}
				c.Dispatch(msg)
			}
		}
	}()

	c.sock = sock

	c.WriteString("NICK " + nick)
	c.WriteString("USER bot * * :" + nick)

	return <-errChan
}

func (c *Connection) WriteString(s string) (int, error) {
	// TODO(scjudd): implement proper WriteString interface
	c.write <- s
	return len(s), nil
}