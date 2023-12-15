package core

import (
	"testing"
)

type splitAddrPortCase struct {
	addrPort string
	addr     string
	port     uint16
	ok       bool
}

// revive:disable:cognitive-complexity
func TestSplitAddrPort(t *testing.T) {
	// revive:enable:cognitive-complexity
	var cases = []splitAddrPortCase{
		{"", "", 0, false},                      // nothing                       BAD
		{":6060", "::", 6060, true},             // no host and port              OK
		{":606.0", "", 0, false},                // no host and bad port          BAD
		{":123456", "", 0, false},               // no host and port out of range BAD
		{"0:6060", "0.0.0.0", 6060, true},       // unspecified IPv4 and port     OK
		{"0.0.0.0:6060", "0.0.0.0", 6060, true}, // unspecified IPv4 and port     OK
		{"[::]:6060", "::", 6060, true},         // unspecified IPv6 and port     OK
		{"::1", "::1", 0, true},                 // IPv6 and no port              OK
		{"[::1]", "::1", 0, true},               // bracketed IPv6 and no port    OK
		{"[::1]:", "", 0, false},                // bracketed IPv6 and empty port BAD
		{"[::1]:port", "", 0, false},            // bracketed IPv6 and bad port   BAD
		{"[::1:1234", "", 0, false},             // incomplete bracketed IPv6     BAD
		{"[::1]:1234", "::1", 1234, true},       // bracketed IPv6 and port       OK
		{"[::1]:123456", "", 0, false},          // IPv6 and port out of range    BAD
	}

	for _, d := range cases {
		a, p, err := SplitAddrPort(d.addrPort)
		if err != nil && !d.ok {
			// failed as expected
			t.Logf("%sSplitAddrPort(%q) -> %q, %q, %#v",
				"", d.addrPort, a.String(), p, err)
		} else if a.String() != d.addr || p != d.port || (err == nil) != d.ok {
			// unexpected result
			t.Errorf("%sSplitAddrPort(%q) -> %q, %q, %#v",
				"FAIL ", d.addrPort, a.String(), p, err)
		} else {
			// expected result
			t.Logf("%sSplitAddrPort(%q) -> %q, %q, %#v",
				"", d.addrPort, a.String(), p, err)
		}
	}
}

type splitHostPortCase struct {
	hostport   string
	host, port string
	ok         bool
}

func TestSplitHostPort(t *testing.T) {
	var cases = []splitHostPortCase{
		{"", "", "", false},                                    // nothing                       BAD
		{":6060", "::", "6060", true},                          // no host and port              OK
		{":606.0", "", "", false},                              // no host and bad port          BAD
		{":123456", "", "", false},                             // no host and port out of range BAD
		{"0:6060", "0.0.0.0", "6060", true},                    // unspecified IPv4 and port     OK
		{"0.0.0.0:6060", "0.0.0.0", "6060", true},              // unspecified IPv4 and port     OK
		{"[::]:6060", "::", "6060", true},                      // unspecified IPv6 and port     OK
		{"localhost", "localhost", "", true},                   // known name and no port        OK
		{"::1", "::1", "", true},                               // IPv6 and no port              OK
		{"[::1]", "::1", "", true},                             // bracketed IPv6 and no port    OK
		{"[::1]:", "", "", false},                              // bracketed IPv6 and empty port BAD
		{"[::1]:port", "", "", false},                          // bracketed IPv6 and bad port   BAD
		{"[::1:1234", "", "", false},                           // incomplete bracketed IPv6     BAD
		{"[::1]:1234", "::1", "1234", true},                    // bracketed IPv6 and port       OK
		{"[::1]:123456", "", "", false},                        // IPv6 and port out of range    BAD
		{"name", "name", "", true},                             // host and no port              OK
		{"name:", "", "", false},                               // host but empty port           BAD
		{"name:1234", "name", "1234", true},                    // simple host and port          OK
		{"name:123.4", "", "", false},                          // bad port                      BAD
		{"name:-123.4", "", "", false},                         // bad port                      BAD
		{"name:123456", "", "", false},                         // port out of range             BAD
		{"name:port", "", "", false},                           // host but bad port             BAD
		{"bad name", "", "", false},                            // bad host no port              BAD
		{"bad..name", "", "", false},                           // bad host no port              BAD
		{".name", "", "", false},                               // bad host no port              BAD
		{"Hello.\u4E16\u754C", "hello.\u4E16\u754C", "", true}, // international name            OK
		{"hello.xn--rhqv96g", "hello.\u4E16\u754C", "", true},  // puny code                     OK
		{"good.name", "good.name", "", true},                   // name with dot and no port     OK
		{":port", "", "", false},                               // no host and bad port          BAD
	}

	for _, d := range cases {
		h, p, err := SplitHostPort(d.hostport)
		if h != d.host || p != d.port || (err == nil) != d.ok {
			t.Errorf("%sSplitHostPort(%q) -> %q, %q, %#v",
				"FAIL ", d.hostport, h, p, err)
		} else {
			t.Logf("%sSplitHostPort(%q) -> %q, %q, %#v",
				"", d.hostport, h, p, err)
		}
	}
}
