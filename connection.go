package irc

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	MaxMessageLen = 512
	MaxNickLen    = 9
)

type Connection struct {
	nick        string
	channels    []string
	sock        net.Conn
	read, write chan string
	MessageDispatcher
}

func NewConnection() *Connection {
	read := make(chan string)
	write := make(chan string)
	dispatcher := NewDispatcher()
	return &Connection{"", nil, nil, read, write, dispatcher}
}

func (c *Connection) Connect(server, nick string) error {
	if len(nick) > MaxNickLen {
		return fmt.Errorf("Nick \"%s\" is too long", nick)
	}
	c.nick = nick

	if c.sock != nil {
		return errors.New("Connection already established")
	}

	if c.MessageDispatcher == nil {
		return errors.New("Connection's MessageDispatcher must be instantiated")
	}

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

	// Read goroutine: Read at most MaxMessageLen bytes off the underlying socket
	// and push complete messages into the Connection.read channel
	go func() {
		br := bufio.NewReaderSize(sock, MaxMessageLen)
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
		bw := bufio.NewWriterSize(sock, MaxMessageLen)
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
				c.Dispatch(msg)
			}
		}
	}()

	c.registerHandlers()

	// Connection initialization: send NICK and USER messages
	if err := c.Nick(nick); err != nil {
		errChan <- err
	}
	c.SendRawMessage(fmt.Sprintf("USER %s * * :%s\r\n", nick, nick))

	return <-errChan
}

func (c *Connection) registerHandlers() {
	// Send PING PONGs
	c.RegisterHandler("PING", func(msg *Message) bool {
		c.SendRawMessage("PONG " + strings.Join(msg.Params, " ") + "\r\n")
		return true
	})

	// Nick state tracking
	c.RegisterHandler("NICK", func(msg *Message) bool {
		if msg.Nick == c.nick {
			c.nick = msg.Params[0]
			return true
		}
		return false
	})

	// Channel state tracking
	c.RegisterHandler("JOIN", func(msg *Message) bool {
		if msg.Nick == c.nick {
			c.channels = append(c.channels, msg.Params[0])
			return true
		}
		return false
	})
	c.RegisterHandler("PART", func(msg *Message) bool {
		if msg.Nick == c.nick {
			for i, ch := range c.channels {
				if ch == msg.Params[0] {
					c.channels = append(c.channels[:i], c.channels[i+1:]...)
				}
			}
			return true
		}
		return false
	})
}

func (c *Connection) SendRawMessage(s string) {
	c.write <- s
}

func (c *Connection) Nick(s string) error {
	if len(s) > MaxNickLen {
		return fmt.Errorf("Nick \"%s\" is too long", s)
	}
	c.SendRawMessage(fmt.Sprintf("NICK %s\r\n", s))
	return nil
}

func (c *Connection) Join(s string) error {
	c.SendRawMessage(fmt.Sprintf("JOIN %s\r\n", s))
	return nil
}

func (c *Connection) Privmsg(target, s string) error {
	c.SendRawMessage(fmt.Sprintf("PRIVMSG %s :%s\r\n", target, s))
	return nil
}
