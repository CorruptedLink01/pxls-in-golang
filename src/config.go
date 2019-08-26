package main

import (
	"io/ioutil"

	"github.com/go-akka/configuration"
)

const (
	// ConfigReferenceFile is the name of the reference config file
	ConfigReferenceFile = "reference.pxls.conf"
	// ConfigFile is the name of the config file
	ConfigFile = "pxls.conf"
)

// ReadConfig reads the HOCON configuration from ./pxls.conf and
// returns a configuration.Config pointer if it was successful,
// or an error if it was not.
func ReadConfig() (conf *configuration.Config, err error) {
	s, err := ioutil.ReadFile(ConfigFile)
	conf = configuration.ParseString(string(s))
	return
}
