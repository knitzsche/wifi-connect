package utils

import "fmt"

// PkgError returns a new instance implementing the Error interface that is intended
// to accept the package name as an input. Callers can wrap this to create Error methods
// for their packages.
func PkgError(pkg string, a ...interface{}) error {
	return fmt.Errorf("%v: %v", pkg, a)
}

// PkgErrorf returns a new instance implementing the Error interface that is intended
// to accept the package name and a format string as inputs. Callers can wrap this to
// create Errorf methods for their packages.
func PkgErrorf(pkg string, format string, a ...interface{}) error {
	return fmt.Errorf(fmt.Sprintf("%v: %v", pkg, format), a...)
}

// PkgSprint returns a formatted string with a project prefix. Callers can wrap this to
// create Sprintf methods for their packages.
func PkgSprint(pkg string, a ...interface{}) string {
	return fmt.Sprintf("%v: %v", pkg, a)
}

// PkgSprintf returns a formatted string with a project prefix. Callers can wrap this to
// create Sprintf methods for their packages.
func PkgSprintf(pkg string, format string, a ...interface{}) string {
	return fmt.Sprintf(fmt.Sprintf("%v: %v", pkg, format), a...)
}
