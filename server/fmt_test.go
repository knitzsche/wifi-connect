package server

import "testing"

const (
	pkg = "server"
)

func TestErrorFormat(t *testing.T) {

	msg := "Whatever message"
	e := Errorf(msg)
	expected := pkg + ": " + msg

	if e.Error() != expected {
		t.Errorf("Error is not well formatted, expected %v but got %v", expected, e.Error())
	}
}
func TestErrorfFormat(t *testing.T) {

	format := "Whatever %v"
	detail := "message"
	e := Errorf(format, detail)

	expected := pkg + ": Whatever message"

	if e.Error() != expected {
		t.Errorf("Error is not well formatted, expected %s but got %s", expected, e.Error())
	}
}
