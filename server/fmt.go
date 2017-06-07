package server

import (
	"launchpad.net/wifi-connect/utils"
)

// Errorf returns a new instance implementing error interface taken a formatted string and
// params as input
func Errorf(format string, a ...interface{}) error {
	return utils.PkgErrorf("server", format, a...)
}

// Sprintf returns a formatted string with project prefix
func Sprintf(format string, a ...interface{}) string {
	return utils.PkgSprintf("server", format, a...)
}
