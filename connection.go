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
	read, write chan string
	MessageDispatcher
}

func NewConnection() *Connection {
	read := make(chan string)
	write := make(chan string)
	dispatcher := NewDispatcher()
	return &Connection{nil, read, write, dispatcher}
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
		c.SendString("PONG " + strings.Join(msg.Params, " ") + "\r\n")
	})

	sock, err := net.Dial("tcp", server)
	if err != nil {
		return err
	}
	c.sock = sock

	errChan := make(chan error)
	shutdownChan := make(chan struct{})
	fatal := func(msg string) {
		errChan <- errors.New(msg)
		close(shutdownChan)
	}

	// Read goroutine: Read at most MaxMsgLen bytes off the underlying socket
	// and push complete messages into the Connection.read channel
	go func() {
		br := bufio.NewReaderSize(sock, MaxMsgLen)
		for {
			select {
			case <-shutdownChan:
				return
			default:
				buff, err := br.ReadString('\n')
				if err != nil {
					// TODO(scjudd): if err is EOF, try reconnecting
					fatal(fmt.Sprintf("read: %s", err))
					return
				}
				c.read <- buff
			}
		}
	}()

	// Write goroutine: Pull complete messages off the Connection.write channel
	// and write them to the underlying socket
	go func() {
		bw := bufio.NewWriterSize(sock, MaxMsgLen)
		for {
			select {
			case <-shutdownChan:
				return
			default:
				buff, ok := <-c.write
				if !ok {
					fatal("write: channel closed")
					return
				}
				_, err := bw.WriteString(buff)
				if err != nil {
					fatal(fmt.Sprintf("write: %s", err))
					return
				}
				bw.Flush()
				log.Printf("\x1b[92;40;1m--> %s\x1b[0m\n", buff[:len(buff)-2]) // remove "\r\n"
			}
		}
	}()

	// Dispatch goroutine: Pull message bytes off the Connection.read channel,
	// create Message structs, and dispatch Messages to registered handlers
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

	c.Nick(nick)
	c.SendString(fmt.Sprintf("USER %s * * :%s\r\n", nick, nick))

	return <-errChan
}

func (c *Connection) SendString(s string) {
	c.write <- s
}

func (c *Connection) Nick(s string) {
	c.SendString(fmt.Sprintf("NICK %s\r\n", s))
}

func (c *Connection) Join(s string) {
	c.SendString(fmt.Sprintf("JOIN %s\r\n", s))
}

func (c *Connection) Privmsg(target, s string) {
	c.SendString(fmt.Sprintf("PRIVMSG %s :%s\r\n", target, s))
}
