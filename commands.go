package main

func (c *Connection) Join(ch string) {
	c.WriteString("JOIN " + ch + "\r\n")
}

func (c *Connection) Privmsg(who, msg string) {
	c.WriteString("PRIVMSG " + who + " :" + msg + "\r\n")
}
