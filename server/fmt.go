package server

import "launchpad.net/wifi-connect/utils"

// Errorf implements the Error interface. It returns an Error with the "server"
// package name automatically inserted as a prefix.
func Errorf(format string, a ...interface{}) error {
	return utils.PkgErrorf("server", format, a...)
}

// Sprintf returns a formatted string with the "server"
// package name automatically inserted as a prefix.
func Sprintf(format string, a ...interface{}) string {
	return utils.PkgSprintf("server", format, a...)
}
