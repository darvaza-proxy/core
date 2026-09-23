package core

import (
	"net"
	"net/netip"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ TestCase = splitAddrPortTestCase{}
	_ TestCase = splitHostPortTestCase{}
	_ TestCase = makeHostPortTestCase{}
	_ TestCase = joinHostPortTestCase{}
	_ TestCase = doMakeHostPortTestCase{}
	_ TestCase = doJoinHostPortTestCase{}
	_ TestCase = ipForHostPortTestCase{}
)

// mustExpected refuses an empty result in an accepting factory. Every
// subject below returns the empty string when it rejects its input, so
// an accepted row declaring one would be a rejected row in disguise.
func mustExpected(s string) string {
	if s == "" {
		panic("accepted row declares an empty result")
	}
	return s
}

type splitAddrPortTestCase struct {
	addr     netip.Addr
	name     string
	addrPort string
	port     uint16
	rejected bool
}

func (tc splitAddrPortTestCase) Name() string {
	return tc.name
}

func (tc splitAddrPortTestCase) Test(t *testing.T) {
	t.Helper()

	addr, port, err := SplitAddrPort(tc.addrPort)

	AssertEqual(t, tc.addr, addr, "address")
	AssertEqual(t, tc.port, port, "port")

	if tc.rejected {
		// A rejection names the whole input.
		rejection := AssertMustErrorAs[*net.AddrError](t, err, "error")
		AssertEqual(t, tc.addrPort, rejection.Addr, "error address")
	} else {
		AssertNoError(t, err, "error")
	}
}

// newSplitAddrPortTestCase declares an accepted input by the address
// and port it splits into; MustParseAddr refuses an empty address.
func newSplitAddrPortTestCase(name, addrPort, addr string, port uint16) splitAddrPortTestCase {
	return splitAddrPortTestCase{
		addr:     netip.MustParseAddr(addr),
		name:     name,
		addrPort: addrPort,
		port:     port,
	}
}

// newSplitAddrPortTestCaseRejected declares an input SplitAddrPort
// rejects, returning its zero values.
func newSplitAddrPortTestCaseRejected(name, addrPort string) splitAddrPortTestCase {
	return splitAddrPortTestCase{
		name:     name,
		addrPort: addrPort,
		rejected: true,
	}
}

func splitAddrPortTestCases() []splitAddrPortTestCase {
	return S(
		// IP addresses
		newSplitAddrPortTestCase("unspecified IPv4", "0.0.0.0:6060", "0.0.0.0", 6060),
		newSplitAddrPortTestCase("IPv6 no port", "::1", "::1", 0),
		newSplitAddrPortTestCase("bracketed IPv6 no port", "[::1]", "::1", 0),
		newSplitAddrPortTestCase("bracketed IPv6 and port", "[::1]:1234", "::1", 1234),
		newSplitAddrPortTestCase("unspecified IPv6", "[::]:6060", "::", 6060),
		newSplitAddrPortTestCase("no host and port", ":6060", "::", 6060),

		// Rejected ports
		newSplitAddrPortTestCaseRejected("bracketed IPv6 bad port", "[::1]:port"),

		// Rejected addresses
		// A valid port but a host that isn't a literal IP forces
		// ParseAddr to fail, covering the non-IP address branch.
		newSplitAddrPortTestCaseRejected("hostname not IP", "name:1234"),
		newSplitAddrPortTestCaseRejected("empty", ""),
	)
}

func TestSplitAddrPort(t *testing.T) {
	RunTestCases(t, splitAddrPortTestCases())
}

type splitHostPortTestCase struct {
	name     string
	hostport string
	host     string
	port     string
	rejected bool
}

func (tc splitHostPortTestCase) Name() string {
	return tc.name
}

func (tc splitHostPortTestCase) Test(t *testing.T) {
	t.Helper()

	host, port, err := SplitHostPort(tc.hostport)

	AssertEqual(t, tc.host, host, "host")
	AssertEqual(t, tc.port, port, "port")

	if tc.rejected {
		// A rejection names the whole input.
		rejection := AssertMustErrorAs[*net.AddrError](t, err, "error")
		AssertEqual(t, tc.hostport, rejection.Addr, "error address")
	} else {
		AssertNoError(t, err, "error")
	}
}

// newSplitHostPortTestCase declares an accepted input by the host and
// port it splits into; the port is empty where the input carries none.
func newSplitHostPortTestCase(name, hostport, host, port string) splitHostPortTestCase {
	return splitHostPortTestCase{
		name:     name,
		hostport: hostport,
		host:     mustExpected(host),
		port:     port,
	}
}

// newSplitHostPortTestCaseRejected declares an input SplitHostPort
// rejects, returning empty strings.
func newSplitHostPortTestCaseRejected(name, hostport string) splitHostPortTestCase {
	return splitHostPortTestCase{
		name:     name,
		hostport: hostport,
		rejected: true,
	}
}

func splitHostPortTestCases() []splitHostPortTestCase {
	return S(
		// Names
		newSplitHostPortTestCase("name", "name", "name", ""),
		newSplitHostPortTestCase("name and port", "name:1234", "name", "1234"),
		newSplitHostPortTestCase("name and padded port", "name:0080", "name", "80"),
		newSplitHostPortTestCase("good name", "good.name", "good.name", ""),
		newSplitHostPortTestCase("international name", "Hello.\u4E16\u754C", "hello.\u4E16\u754C", ""),
		newSplitHostPortTestCase("puny code", "hello.xn--rhqv96g", "hello.\u4E16\u754C", ""),

		// IP addresses
		newSplitHostPortTestCase("unspecified IPv4", "0.0.0.0:6060", "0.0.0.0", "6060"),
		newSplitHostPortTestCase("IPv6 no port", "::1", "::1", ""),
		newSplitHostPortTestCase("bracketed IPv6 no port", "[::1]", "::1", ""),
		newSplitHostPortTestCase("bracketed IPv6 and port", "[::1]:1234", "::1", "1234"),
		newSplitHostPortTestCase("unspecified IPv6", "[::]:6060", "::", "6060"),
		newSplitHostPortTestCase("no host and port", ":6060", "::", "6060"),

		// Rejected ports
		newSplitHostPortTestCaseRejected("name empty port", "name:"),
		newSplitHostPortTestCaseRejected("name bad port", "name:123.4"),
		newSplitHostPortTestCaseRejected("name negative port", "name:-123.4"),
		newSplitHostPortTestCaseRejected("name port out of range", "name:123456"),
		newSplitHostPortTestCaseRejected("name non-numeric port", "name:port"),
		newSplitHostPortTestCaseRejected("bracketed IPv6 empty port", "[::1]:"),

		// Rejected hosts
		newSplitHostPortTestCaseRejected("bad hostname spaces", "bad name"),
		newSplitHostPortTestCaseRejected("bad hostname dots", "bad..name"),
		newSplitHostPortTestCaseRejected("bad hostname leading dot", ".name"),
		newSplitHostPortTestCaseRejected("incomplete bracketed IPv6", "[::1:1234"),
		// Trailing garbage after `]` exercises the default branch of
		// splitHostPortBracketed.
		newSplitHostPortTestCaseRejected("bracketed IPv6 trailing garbage", "[::1]x"),
		newSplitHostPortTestCaseRejected("empty", ""),
	)
}

func TestSplitHostPort(t *testing.T) {
	RunTestCases(t, splitHostPortTestCases())
}

type makeHostPortTestCase struct {
	name        string
	hostPort    string
	expected    string
	defaultPort uint16
	rejected    bool
}

func (tc makeHostPortTestCase) Name() string {
	return tc.name
}

func (tc makeHostPortTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := MakeHostPort(tc.hostPort, tc.defaultPort)

	AssertEqual(t, tc.expected, got, "host:port")

	if tc.rejected {
		// A rejection names the whole input.
		rejection := AssertMustErrorAs[*net.AddrError](t, err, "error")
		AssertEqual(t, tc.hostPort, rejection.Addr, "error address")
	} else {
		AssertNoError(t, err, "error")
	}
}

// newMakeHostPortTestCase declares an accepted input by the host:port
// it produces.
func newMakeHostPortTestCase(name, hostPort string, defaultPort uint16, expected string) makeHostPortTestCase {
	return makeHostPortTestCase{
		name:        name,
		hostPort:    hostPort,
		expected:    mustExpected(expected),
		defaultPort: defaultPort,
	}
}

// newMakeHostPortTestCaseRejected declares an input MakeHostPort
// rejects, returning the empty string.
func newMakeHostPortTestCaseRejected(name, hostPort string, defaultPort uint16) makeHostPortTestCase {
	return makeHostPortTestCase{
		name:        name,
		hostPort:    hostPort,
		defaultPort: defaultPort,
		rejected:    true,
	}
}

func makeHostPortTestCases() []makeHostPortTestCase {
	return S(
		// Valid cases with IP addresses
		newMakeHostPortTestCase("IPv6 bracketed no port", "[::1]", 0, "::1"),
		newMakeHostPortTestCase("IPv6 unbracketed default port", "::1", 8080, "[::1]:8080"),

		// Valid cases with hostnames
		newMakeHostPortTestCase("FQDN default port", "example.com", 443, "example.com:443"),
		newMakeHostPortTestCase("FQDN explicit port", "example.com:80", 443, "example.com:80"),
		newMakeHostPortTestCase("FQDN padded port", "example.com:0080", 443, "example.com:80"),
		newMakeHostPortTestCase("FQDN no port", "example.com", 0, "example.com"),

		// Invalid cases
		newMakeHostPortTestCaseRejected("empty input", "", 8080),
		newMakeHostPortTestCaseRejected("invalid hostname", "invalid host", 8080),
		newMakeHostPortTestCaseRejected("port 0 not allowed", "example.com:0", 8080),
		newMakeHostPortTestCaseRejected("port 0 padded", "example.com:00", 8080),
		// Port 0 in the input is rejected, not read as portless: the
		// same default that accepts "example.com" does not rescue it.
		newMakeHostPortTestCaseRejected("port 0 without default", "example.com:0", 0),
		newMakeHostPortTestCaseRejected("invalid port", "example.com:invalid", 8080),
		// Port 0 is judged after the split has cleaned the host and the
		// port, and the error names the input.
		newMakeHostPortTestCaseRejected("port 0 cleaned name", "Example.com:0", 8080),
		newMakeHostPortTestCaseRejected("port 0 no host", ":0", 8080),
	)
}

func TestMakeHostPort(t *testing.T) {
	RunTestCases(t, makeHostPortTestCases())
}

type joinHostPortTestCase struct {
	name     string
	host     string
	port     string
	expected string
	errAddr  string
	rejected bool
}

func (tc joinHostPortTestCase) Name() string {
	return tc.name
}

func (tc joinHostPortTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := JoinHostPort(tc.host, tc.port)

	AssertEqual(t, tc.expected, got, "host:port")

	if tc.rejected {
		rejection := AssertMustErrorAs[*net.AddrError](t, err, "error")
		AssertEqual(t, tc.errAddr, rejection.Addr, "error address")
	} else {
		AssertNoError(t, err, "error")
	}
}

// newJoinHostPortTestCase declares an accepted pair by the host:port
// it joins into.
func newJoinHostPortTestCase(name, host, port, expected string) joinHostPortTestCase {
	return joinHostPortTestCase{
		name:     name,
		host:     host,
		port:     port,
		expected: mustExpected(expected),
	}
}

// newJoinHostPortTestCaseRejected declares a pair JoinHostPort rejects,
// returning the empty string, by the address its error names.
func newJoinHostPortTestCaseRejected(name, host, port, errAddr string) joinHostPortTestCase {
	return joinHostPortTestCase{
		name:     name,
		host:     host,
		port:     port,
		errAddr:  errAddr,
		rejected: true,
	}
}

func joinHostPortTestCases() []joinHostPortTestCase {
	return S(
		// Valid cases with IP addresses
		newJoinHostPortTestCase("IPv6 with port", "::1", "8080", "[::1]:8080"),
		newJoinHostPortTestCase("IPv6 no port", "::1", "", "::1"),

		// Valid cases with hostnames
		newJoinHostPortTestCase("FQDN with port", "example.com", "443", "example.com:443"),
		newJoinHostPortTestCase("FQDN no port", "example.com", "", "example.com"),
		newJoinHostPortTestCase("padded port", "example.com", "0080", "example.com:80"),

		// Port 0 joins, where MakeHostPort rejects it in its input.
		newJoinHostPortTestCase("port 0 valid", "example.com", "0", "example.com:0"),

		// Invalid cases
		// A rejected host is named alone, the port not having been read.
		newJoinHostPortTestCaseRejected("empty host", "", "8080", ""),
		newJoinHostPortTestCaseRejected("invalid hostname", "invalid host", "8080", "invalid host"),

		// A rejected port is named with the host before it.
		newJoinHostPortTestCaseRejected("negative port", "example.com", "-1", "example.com:-1"),
	)
}

func TestJoinHostPort(t *testing.T) {
	RunTestCases(t, joinHostPortTestCases())
}

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
		expected:    mustExpected(expected),
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
		expected: mustExpected(expected),
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
