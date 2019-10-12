package main

import (
	"testing"
)

const errorFormat = "expected:%v, actual:%v"

func TestParseCookieNameAndDomain(t *testing.T) {
	s := "__cfduid=01ab; expires=Sun, 20-Sep-20 07:42:35 GMT; path=/; domain=.2ch.net; HttpOnly"
	actual := parseCookieNameAndDomain(s)
	expectend := "__cfduid;.2ch.net"
	if actual != expectend {
		t.Errorf(errorFormat, expectend, actual)
	}
}

func TestParseCookieNameAndDomain2(t *testing.T) {
	s := "__cfduid=01ab; expires=Sun, 20-Sep-20 07:42:35 GMT; path=/; HttpOnly"
	actual := parseCookieNameAndDomain(s)
	expectend := "__cfduid;"
	if actual != expectend {
		t.Errorf(errorFormat, expectend, actual)
	}
}
