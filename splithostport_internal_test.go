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
	t.Helper()

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
		expected:    MustNotEmptyForTest(expected),
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
		// Valid cases with explicit port
		newDoMakeHostPortTestCase("explicit port used", "example.com", "8080", 9000, "example.com:8080"),
		newDoMakeHostPortTestCase("explicit port padded", "example.com", "0080", 9000, "example.com:80"),

		// Valid cases with default port
		newDoMakeHostPortTestCase("default port used", "example.com", "", 8080, "example.com:8080"),

		// Valid cases with no port
		newDoMakeHostPortTestCase("no port hostname", "example.com", "", 0, "example.com"),
		newDoMakeHostPortTestCase("no port IPv6", "[::1]", "", 0, "[::1]"),

		// Invalid cases
		newDoMakeHostPortTestCaseRejected("port 0 not allowed", "example.com", "0", 8080),
		// MakeHostPort never hands over a padded port, the split having
		// made it canonical, so only a direct call reaches this.
		newDoMakeHostPortTestCaseRejected("port 0 padded", "example.com", "00", 8080),
		// MakeHostPort never hands over a bad port, the split having
		// refused it, so only a direct call reaches this.
		newDoMakeHostPortTestCaseRejected("bad port", "example.com", "invalid", 8080),
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
	t.Helper()

	got, err := doJoinHostPort(tc.host, tc.port)

	AssertEqual(t, tc.expected, got, "host:port")

	if tc.rejected {
		// A rejection names the host and port as given.
		rejection := AssertMustErrorAs[*net.AddrError](t, err, "error")
		AssertEqual(t, tc.host+":"+tc.port, rejection.Addr, "error address")
	} else {
		AssertNoError(t, err, "error")
	}
}

// newDoJoinHostPortTestCase declares an accepted pair by the host:port
// it joins into.
func newDoJoinHostPortTestCase(name, host, port, expected string) doJoinHostPortTestCase {
	return doJoinHostPortTestCase{
		name:     name,
		host:     host,
		port:     port,
		expected: MustNotEmptyForTest(expected),
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
		// Valid cases
		newDoJoinHostPortTestCase("valid hostname", "example.com", "8080", "example.com:8080"),
		newDoJoinHostPortTestCase("padded port", "example.com", "0080", "example.com:80"),

		// Invalid cases
		newDoJoinHostPortTestCaseRejected("port out of range high", "example.com", "65536"),
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
	t.Helper()

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
		// IPv4 addresses (should not be bracketed)
		newIPForHostPortTestCase("IPv4 localhost", "127.0.0.1", "127.0.0.1"),

		// IPv6 addresses (should be bracketed)
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
