package utils

import (
	"testing"
)

func TestPkgErrorFormat(t *testing.T) {

	msg := "Whatever message"
	e := PkgErrorf("package", msg)
	expected := "package: " + msg

	if e.Error() != expected {
		t.Errorf("Error is not well formatted, expected %v but got %v", expected, e.Error())
	}
}
func TestPkgErrorfFormat(t *testing.T) {

	format := "Whatever %v"
	detail := "message"
	e := PkgErrorf("package", format, detail)

	expected := "package: Whatever message"

	if e.Error() != expected {
		t.Errorf("Error is not well formatted, expected %s but got %s", expected, e.Error())
	}
}

func TestPkgSprintFormat(t *testing.T) {

	msg := "Whatever message"
	e := PkgSprintf("package", msg)
	expected := "package: " + msg

	if e != expected {
		t.Errorf("String is not well formatted, expected %s but got %s", expected, e)
	}
}

func TestPkgSprintfFormat(t *testing.T) {

	format := "Whatever %v"
	detail := "message"
	e := PkgSprintf("package", format, detail)

	expected := "package: Whatever message"

	if e != expected {
		t.Errorf("String is not well formatted, expected %s but got %s", expected, e)
	}
}
