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
	read, write chan []byte
	MessageDispatcher
}

func NewConnection() *Connection {
	read := make(chan []byte)
	write := make(chan []byte)
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
		c.WriteString("PONG " + strings.Join(msg.Params, " ") + "\r\n")
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

	// Read goroutine: Read at most MaxMsgLen bytes off the underlying socket
	// and push complete messages into the Connection.read channel
	go func() {
		br := bufio.NewReaderSize(sock, MaxMsgLen)
		for {
			select {
			case <-shutdownChan:
				return
			default:
				buff, err := br.ReadBytes('\n')
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
				_, err := bw.Write(buff)
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
				msg := parseMessage(string(<-c.read))
				// TODO(scjudd): proper prefix parsing, so someone with nick botbot won't get ignored
				if strings.Index(msg.Prefix, nick) == 0 {
					continue // ignore our own messages
				}
				c.Dispatch(msg)
			}
		}
	}()

	c.sock = sock

	c.WriteString("NICK " + nick + "\r\n")
	c.WriteString("USER bot * * :" + nick + "\r\n")

	return <-errChan
}

func (c *Connection) Write(b []byte) (int, error) {
	// TODO(scjudd): implement proper Write interface
	c.write <- b
	return len(b), nil
}

func (c *Connection) WriteString(s string) (int, error) {
	return c.Write([]byte(s))
}
