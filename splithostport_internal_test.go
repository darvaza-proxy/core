package core

import (
	"net"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = doMakeHostPortTestCase{}
	_ TestCase = doJoinHostPortTestCase{}
	_ TestCase = ipForHostPortTestCase{}
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
		newDoMakeHostPortTestCase("default port IPv6", "[::1]", "", 8080, "[::1]:8080"),

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

type doJoinHostPortTestCase struct {
	name     string
	host     string
	port     string
	expected string
	rejected bool
}

func (tc doJoinHostPortTestCase) Name() string {
	return tc.name
}

func (tc doJoinHostPortTestCase) Test(t *testing.T) {
	got, err := doJoinHostPort(tc.host, tc.port)

	AssertEqual(t, tc.expected, got, "host:port")

	if tc.rejected {
		// A rejection names the host and port as given.
		rejection := AssertMustErrorAs[*net.AddrError](t, err, "error")
		AssertEqual(t, tc.host+":"+tc.port, rejection.Addr, "error address")
		return
	}

	AssertNoError(t, err, "error")
}

// newDoJoinHostPortTestCase declares an accepted pair by the host:port
// it joins into.
func newDoJoinHostPortTestCase(name, host, port, expected string) doJoinHostPortTestCase {
	return doJoinHostPortTestCase{
		name:     name,
		host:     host,
		port:     port,
		expected: MustNotZeroForTest(expected),
	}
}

// newDoJoinHostPortTestCaseRejected declares a pair doJoinHostPort
// rejects, returning the empty string.
func newDoJoinHostPortTestCaseRejected(name, host, port string) doJoinHostPortTestCase {
	return doJoinHostPortTestCase{
		name:     name,
		host:     host,
		port:     port,
		rejected: true,
	}
}

func doJoinHostPortTestCases() []doJoinHostPortTestCase {
	return S(
		// The port joins in canonical form.
		newDoJoinHostPortTestCase("plain port", "example.com", "8080", "example.com:8080"),
		newDoJoinHostPortTestCase("padded port", "example.com", "0080", "example.com:80"),

		// A port beyond 65535 is rejected.
		newDoJoinHostPortTestCaseRejected("port 65536", "example.com", "65536"),
	)
}

func TestDoJoinHostPort(t *testing.T) {
	RunTestCases(t, doJoinHostPortTestCases())
}

type ipForHostPortTestCase struct {
	name     string
	input    string
	expected string
}

func (tc ipForHostPortTestCase) Name() string {
	return tc.name
}

func (tc ipForHostPortTestCase) Test(t *testing.T) {
	addr, err := ParseAddr(tc.input)
	AssertMustNoError(t, err, "ParseAddr")

	AssertEqual(t, tc.expected, ipForHostPort(addr), "host form")
}

func newIPForHostPortTestCase(name, input, expected string) ipForHostPortTestCase {
	return ipForHostPortTestCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func ipForHostPortTestCases() []ipForHostPortTestCase {
	return S(
		// An IPv4 address comes back as it is.
		newIPForHostPortTestCase("IPv4 localhost", "127.0.0.1", "127.0.0.1"),

		// An IPv6 address comes back bracketed, a mapped one included.
		newIPForHostPortTestCase("IPv6 localhost", "::1", "[::1]"),
		newIPForHostPortTestCase("IPv6 mapped", "::ffff:192.0.2.1", "[::ffff:192.0.2.1]"),
	)
}

func TestIPForHostPort(t *testing.T) {
	RunTestCases(t, ipForHostPortTestCases())
}

// TestAddrErr states what addrErr hands its callers: a *net.AddrError
// carrying its two arguments in the two fields.
func TestAddrErr(t *testing.T) {
	err := addrErr("invalid.address", "test error")

	got := AssertMustErrorAs[*net.AddrError](t, err, "addrErr")
	AssertEqual(t, "invalid.address", got.Addr, "Addr")
	AssertEqual(t, "test error", got.Err, "Err")
}
