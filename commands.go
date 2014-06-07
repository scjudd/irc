package main

func (c *Connection) Join(ch string) {
	c.SendString("JOIN " + ch + "\r\n")
}

func (c *Connection) Privmsg(who, msg string) {
	c.SendString("PRIVMSG " + who + " :" + msg + "\r\n")
}
