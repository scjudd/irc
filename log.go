package irc

import (
	"io/ioutil"
	"log"
	"os"
)

func init() {
	DisableLogging()
}

func DisableLogging() {
	log.SetOutput(ioutil.Discard)
}

func EnableLogging() {
	log.SetOutput(os.Stdout)
}
