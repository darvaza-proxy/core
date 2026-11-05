package core

import (
	"net"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = doMakeHostPortTestCase{}
	_ TestCase = joinHostPortUnsafeTestCase{}
)

type doMakeHostPortTestCase struct {
	name        string
	host        string
	port        string
	expected    string
	defaultPort uint16
	rejected    bool
}

func (tc doMakeHostPortTestCase) Name() string {
	return tc.name
}

func (tc doMakeHostPortTestCase) Test(t *testing.T) {
	got, ok := doMakeHostPort(tc.host, tc.port, tc.defaultPort)

	AssertEqual(t, tc.expected, got, "host:port")
	AssertEqual(t, !tc.rejected, ok, "accepted")
}

// newDoMakeHostPortTestCase declares an accepted input by the host:port
// it produces.
func newDoMakeHostPortTestCase(name, host, port string, defaultPort uint16,
	expected string) doMakeHostPortTestCase {
	return doMakeHostPortTestCase{
		name:        name,
		host:        host,
		port:        port,
		expected:    MustNotZeroForTest(expected),
		defaultPort: defaultPort,
	}
}

// newDoMakeHostPortTestCaseRejected declares an input doMakeHostPort
// rejects, returning the empty string and false.
func newDoMakeHostPortTestCaseRejected(name, host, port string, defaultPort uint16) doMakeHostPortTestCase {
	return doMakeHostPortTestCase{
		name:        name,
		host:        host,
		port:        port,
		defaultPort: defaultPort,
		rejected:    true,
	}
}

func doMakeHostPortTestCases() []doMakeHostPortTestCase {
	return S(
		// An explicit port wins over the default, in canonical form.
		newDoMakeHostPortTestCase("explicit port", "example.com", "8080", 9000, "example.com:8080"),
		// MakeHostPort passes on the port the split made canonical, so
		// this row hands doMakeHostPort a padded one directly.
		newDoMakeHostPortTestCase("explicit port padded", "example.com", "0080", 9000, "example.com:80"),

		// Without one, the default port is used.
		newDoMakeHostPortTestCase("default port", "example.com", "", 8080, "example.com:8080"),
		newDoMakeHostPortTestCase("default port IPv6", "::1", "", 8080, "[::1]:8080"),

		// Without either, the host comes back portless.
		newDoMakeHostPortTestCase("no port hostname", "example.com", "", 0, "example.com"),

		// Port 0 is rejected, in any spelling, and so is a bad port.
		newDoMakeHostPortTestCaseRejected("port 0", "example.com", "0", 8080),
		// As "explicit port padded" does, this row hands doMakeHostPort
		// a padded port directly.
		newDoMakeHostPortTestCaseRejected("port 0 padded", "example.com", "00", 8080),
		// This row hands doMakeHostPort the bad port that the split
		// refuses before MakeHostPort sees it.
		newDoMakeHostPortTestCaseRejected("word port", "example.com", "invalid", 8080),
	)
}

func TestDoMakeHostPort(t *testing.T) {
	RunTestCases(t, doMakeHostPortTestCases())
}

type joinHostPortUnsafeTestCase struct {
	name     string
	host     string
	port     string
	expected string
}

func (tc joinHostPortUnsafeTestCase) Name() string {
	return tc.name
}

func (tc joinHostPortUnsafeTestCase) Test(t *testing.T) {
	AssertEqual(t, tc.expected, joinHostPortUnsafe(tc.host, tc.port), "host:port")
}

// newJoinHostPortUnsafeTestCase declares a pair by the string
// joinHostPortUnsafe joins it into.
func newJoinHostPortUnsafeTestCase(name, host, port, expected string) joinHostPortUnsafeTestCase {
	return joinHostPortUnsafeTestCase{
		name:     name,
		host:     host,
		port:     port,
		expected: expected,
	}
}

func joinHostPortUnsafeTestCases() []joinHostPortUnsafeTestCase {
	return S(
		// A port follows the host after a ':'.
		newJoinHostPortUnsafeTestCase("name and port", "example.com", "80", "example.com:80"),
		newJoinHostPortUnsafeTestCase("IPv4 and port", "127.0.0.1", "80", "127.0.0.1:80"),

		// A host holding a ':' is bracketed before the port, a mapped
		// or zoned IPv6 address included.
		newJoinHostPortUnsafeTestCase("IPv6 and port", "::1", "80", "[::1]:80"),
		newJoinHostPortUnsafeTestCase("IPv6 mapped and port", "::ffff:192.0.2.1", "80", "[::ffff:192.0.2.1]:80"),
		newJoinHostPortUnsafeTestCase("IPv6 zone and port", "fe80::1%eth0", "80", "[fe80::1%eth0]:80"),

		// Without a port, the host comes back alone, unbracketed.
		newJoinHostPortUnsafeTestCase("name no port", "example.com", "", "example.com"),
		newJoinHostPortUnsafeTestCase("IPv6 no port", "::1", "", "::1"),

		// Neither the host nor the port is validated.
		newJoinHostPortUnsafeTestCase("invalid host and port", "bad host", "port", "bad host:port"),
		newJoinHostPortUnsafeTestCase("empty host and port", "", "", ""),
	)
}

func TestJoinHostPortUnsafe(t *testing.T) {
	RunTestCases(t, joinHostPortUnsafeTestCases())
}

// TestAddrErr states what addrErr hands its callers: a *net.AddrError
// carrying its two arguments in the two fields.
func TestAddrErr(t *testing.T) {
	err := addrErr("invalid.address", "test error")

	got := AssertMustErrorAs[*net.AddrError](t, err, "addrErr")
	AssertEqual(t, "invalid.address", got.Addr, "Addr")
	AssertEqual(t, "test error", got.Err, "Err")
}
