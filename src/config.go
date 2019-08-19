package main

import (
	"io/ioutil"

	"github.com/go-akka/configuration"
)

// ReadConfig reads the HOCON configuration from ./pxls.conf and
// returns a configuration.Config pointer if it was successful,
// or an error if it was not.
func ReadConfig() (conf *configuration.Config, err error) {
	s, err := ioutil.ReadFile("pxls.conf")
	conf = configuration.ParseString(string(s))
	return
}
