package main

func (c *Connection) Join(ch string) {
	c.SendRaw([]byte("JOIN " + ch + "\r\n"))
}

func (c *Connection) Privmsg(who, msg string) {
	c.SendRaw([]byte("PRIVMSG " + who + " :" + msg + "\r\n"))
}
