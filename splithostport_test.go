package core

import (
	"net"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ TestCase = splitAddrPortCase{}
var _ TestCase = splitHostPortCase{}
var _ TestCase = makeHostPortCase{}
var _ TestCase = joinHostPortCase{}
var _ TestCase = doMakeHostPortCase{}
var _ TestCase = doJoinHostPortCase{}
var _ TestCase = ipForHostPortCase{}
var _ TestCase = addrErrCase{}

type splitAddrPortCase struct {
	name     string
	addrPort string
	addr     string
	port     uint16
	ok       bool
}

func (tc splitAddrPortCase) Name() string {
	return tc.name
}

func (tc splitAddrPortCase) Test(t *testing.T) {
	t.Helper()

	a, p, err := SplitAddrPort(tc.addrPort)
	if err != nil && !tc.ok {
		// failed as expected
		t.Logf("SplitAddrPort(%q) -> %q, %q, %#v", tc.addrPort, a.String(), p, err)
	} else if a.String() != tc.addr || p != tc.port || (err == nil) != tc.ok {
		// unexpected result
		t.Errorf("SplitAddrPort(%q) -> %q, %q, %#v", tc.addrPort, a.String(), p, err)
	} else {
		// expected result
		t.Logf("SplitAddrPort(%q) -> %q, %q, %#v", tc.addrPort, a.String(), p, err)
	}
}

func newSplitAddrPortCase(name, addrPort, addr string, port uint16, ok bool) splitAddrPortCase {
	return splitAddrPortCase{
		name:     name,
		addrPort: addrPort,
		addr:     addr,
		port:     port,
		ok:       ok,
	}
}

func splitAddrPortTestCases() []splitAddrPortCase {
	return S(
		newSplitAddrPortCase("empty", "", "", 0, false),
		newSplitAddrPortCase("no host and port", ":6060", "::", 6060, true),
		newSplitAddrPortCase("no host and bad port", ":606.0", "", 0, false),
		newSplitAddrPortCase("no host and port out of range", ":123456", "", 0, false),
		newSplitAddrPortCase("unspecified IPv4 short", "0:6060", "0.0.0.0", 6060, true),
		newSplitAddrPortCase("unspecified IPv4", "0.0.0.0:6060", "0.0.0.0", 6060, true),
		newSplitAddrPortCase("unspecified IPv6", "[::]:6060", "::", 6060, true),
		newSplitAddrPortCase("IPv6 no port", "::1", "::1", 0, true),
		newSplitAddrPortCase("bracketed IPv6 no port", "[::1]", "::1", 0, true),
		newSplitAddrPortCase("bracketed IPv6 empty port", "[::1]:", "", 0, false),
		newSplitAddrPortCase("bracketed IPv6 bad port", "[::1]:port", "", 0, false),
		newSplitAddrPortCase("incomplete bracketed IPv6", "[::1:1234", "", 0, false),
		newSplitAddrPortCase("bracketed IPv6 and port", "[::1]:1234", "::1", 1234, true),
		newSplitAddrPortCase("IPv6 port out of range", "[::1]:123456", "", 0, false),
	)
}

func TestSplitAddrPort(t *testing.T) {
	RunTestCases(t, splitAddrPortTestCases())
}

type splitHostPortCase struct {
	name       string
	hostport   string
	host, port string
	ok         bool
}

func (tc splitHostPortCase) Name() string {
	return tc.name
}

func (tc splitHostPortCase) Test(t *testing.T) {
	t.Helper()

	h, p, err := SplitHostPort(tc.hostport)
	if h != tc.host || p != tc.port || (err == nil) != tc.ok {
		t.Errorf("SplitHostPort(%q) -> %q, %q, %#v", tc.hostport, h, p, err)
	} else {
		t.Logf("SplitHostPort(%q) -> %q, %q, %#v", tc.hostport, h, p, err)
	}
}

func newSplitHostPortCase(name, hostport, host, port string, ok bool) splitHostPortCase {
	return splitHostPortCase{
		name:     name,
		hostport: hostport,
		host:     host,
		port:     port,
		ok:       ok,
	}
}

func splitHostPortTestCases() []splitHostPortCase {
	return S(
		newSplitHostPortCase("empty", "", "", "", false),
		newSplitHostPortCase("no host and port", ":6060", "::", "6060", true),
		newSplitHostPortCase("no host and bad port", ":606.0", "", "", false),
		newSplitHostPortCase("no host and port out of range", ":123456", "", "", false),
		newSplitHostPortCase("unspecified IPv4 short", "0:6060", "0.0.0.0", "6060", true),
		newSplitHostPortCase("unspecified IPv4", "0.0.0.0:6060", "0.0.0.0", "6060", true),
		newSplitHostPortCase("unspecified IPv6", "[::]:6060", "::", "6060", true),
		newSplitHostPortCase("hostname", "localhost", "localhost", "", true),
		newSplitHostPortCase("IPv6 no port", "::1", "::1", "", true),
		newSplitHostPortCase("bracketed IPv6 no port", "[::1]", "::1", "", true),
		newSplitHostPortCase("bracketed IPv6 empty port", "[::1]:", "", "", false),
		newSplitHostPortCase("bracketed IPv6 bad port", "[::1]:port", "", "", false),
		newSplitHostPortCase("incomplete bracketed IPv6", "[::1:1234", "", "", false),
		newSplitHostPortCase("bracketed IPv6 and port", "[::1]:1234", "::1", "1234", true),
		newSplitHostPortCase("IPv6 port out of range", "[::1]:123456", "", "", false),
		newSplitHostPortCase("name", "name", "name", "", true),
		newSplitHostPortCase("name empty port", "name:", "", "", false),
		newSplitHostPortCase("name and port", "name:1234", "name", "1234", true),
		newSplitHostPortCase("name bad port", "name:123.4", "", "", false),
		newSplitHostPortCase("name negative port", "name:-123.4", "", "", false),
		newSplitHostPortCase("name port out of range", "name:123456", "", "", false),
		newSplitHostPortCase("name non-numeric port", "name:port", "", "", false),
		newSplitHostPortCase("bad hostname spaces", "bad name", "", "", false),
		newSplitHostPortCase("bad hostname dots", "bad..name", "", "", false),
		newSplitHostPortCase("bad hostname leading dot", ".name", "", "", false),
		newSplitHostPortCase("international name", "Hello.\u4E16\u754C", "hello.\u4E16\u754C", "", true),
		newSplitHostPortCase("puny code", "hello.xn--rhqv96g", "hello.\u4E16\u754C", "", true),
		newSplitHostPortCase("good name", "good.name", "good.name", "", true),
		newSplitHostPortCase("no host bad port", ":port", "", "", false),
	)
}

func TestSplitHostPort(t *testing.T) {
	RunTestCases(t, splitHostPortTestCases())
}

type makeHostPortCase struct {
	name        string
	hostPort    string
	expected    string
	ok          bool
	defaultPort uint16
}

func (tc makeHostPortCase) Name() string {
	return tc.name
}

func (tc makeHostPortCase) Test(t *testing.T) {
	t.Helper()

	result, err := MakeHostPort(tc.hostPort, tc.defaultPort)
	if (err == nil) != tc.ok || (tc.ok && result != tc.expected) {
		t.Errorf("MakeHostPort(%q, %d) -> %q, %v; expected %q, ok=%v",
			tc.hostPort, tc.defaultPort, result, err, tc.expected, tc.ok)
	} else if tc.ok {
		t.Logf("MakeHostPort(%q, %d) -> %q ✓", tc.hostPort, tc.defaultPort, result)
	} else {
		t.Logf("MakeHostPort(%q, %d) -> error: %v ✓", tc.hostPort, tc.defaultPort, err)
	}
}

func newMakeHostPortCase(name, hostPort string, defaultPort uint16, expected string, ok bool) makeHostPortCase {
	return makeHostPortCase{
		name:        name,
		hostPort:    hostPort,
		expected:    expected,
		ok:          ok,
		defaultPort: defaultPort,
	}
}

func makeHostPortTestCases() []makeHostPortCase {
	return S(
		// Valid cases with IP addresses
		newMakeHostPortCase("IPv4 default port", "192.168.1.1", 8080, "192.168.1.1:8080", true),
		newMakeHostPortCase("IPv4 explicit port", "192.168.1.1:9000", 8080, "192.168.1.1:9000", true),
		newMakeHostPortCase("IPv4 no port", "192.168.1.1", 0, "192.168.1.1", true),
		newMakeHostPortCase("IPv6 bracketed default port", "[::1]", 8080, "[::1]:8080", true),
		newMakeHostPortCase("IPv6 bracketed explicit port", "[::1]:9000", 8080, "[::1]:9000", true),
		newMakeHostPortCase("IPv6 bracketed no port", "[::1]", 0, "::1", true),
		newMakeHostPortCase("IPv6 unbracketed default port", "::1", 8080, "[::1]:8080", true),
		newMakeHostPortCase("IPv6 unbracketed no port", "::1", 0, "::1", true),

		// Valid cases with hostnames
		newMakeHostPortCase("hostname default port", "localhost", 8080, "localhost:8080", true),
		newMakeHostPortCase("hostname explicit port", "localhost:9000", 8080, "localhost:9000", true),
		newMakeHostPortCase("hostname no port", "localhost", 0, "localhost", true),
		newMakeHostPortCase("FQDN default port", "example.com", 443, "example.com:443", true),
		newMakeHostPortCase("FQDN explicit port", "example.com:80", 443, "example.com:80", true),

		// Invalid cases
		newMakeHostPortCase("empty input", "", 8080, "", false),
		newMakeHostPortCase("invalid hostname", "invalid host", 8080, "", false),
		newMakeHostPortCase("port 0 not allowed", "example.com:0", 8080, "", false),
		newMakeHostPortCase("port out of range", "example.com:99999", 8080, "", false),
		newMakeHostPortCase("invalid port", "example.com:invalid", 8080, "", false),
		newMakeHostPortCase("malformed IPv6", "[::1", 8080, "", false),
		newMakeHostPortCase("IPv6 invalid port", "[::1]:invalid", 8080, "", false),
	)
}

func TestMakeHostPort(t *testing.T) {
	RunTestCases(t, makeHostPortTestCases())
}

type joinHostPortCase struct {
	name     string
	host     string
	port     string
	expected string
	ok       bool
}

func (tc joinHostPortCase) Name() string {
	return tc.name
}

func (tc joinHostPortCase) Test(t *testing.T) {
	t.Helper()

	result, err := JoinHostPort(tc.host, tc.port)
	if (err == nil) != tc.ok || (tc.ok && result != tc.expected) {
		t.Errorf("JoinHostPort(%q, %q) -> %q, %v; expected %q, ok=%v",
			tc.host, tc.port, result, err, tc.expected, tc.ok)
	} else if tc.ok {
		t.Logf("JoinHostPort(%q, %q) -> %q ✓", tc.host, tc.port, result)
	} else {
		t.Logf("JoinHostPort(%q, %q) -> error: %v ✓", tc.host, tc.port, err)
	}
}

func newJoinHostPortCase(name, host, port, expected string, ok bool) joinHostPortCase {
	return joinHostPortCase{
		name:     name,
		host:     host,
		port:     port,
		expected: expected,
		ok:       ok,
	}
}

func joinHostPortTestCases() []joinHostPortCase {
	return S(
		// Valid cases with IP addresses
		newJoinHostPortCase("IPv4 with port", "192.168.1.1", "8080", "192.168.1.1:8080", true),
		newJoinHostPortCase("IPv4 no port", "192.168.1.1", "", "192.168.1.1", true),
		newJoinHostPortCase("IPv6 with port", "::1", "8080", "[::1]:8080", true),
		newJoinHostPortCase("IPv6 no port", "::1", "", "::1", true),
		newJoinHostPortCase("IPv6 long with port", "2001:db8::1", "9000", "[2001:db8::1]:9000", true),
		newJoinHostPortCase("IPv6 long no port", "2001:db8::1", "", "2001:db8::1", true),

		// Valid cases with hostnames
		newJoinHostPortCase("hostname with port", "localhost", "8080", "localhost:8080", true),
		newJoinHostPortCase("hostname no port", "localhost", "", "localhost", true),
		newJoinHostPortCase("FQDN with port", "example.com", "443", "example.com:443", true),
		newJoinHostPortCase("FQDN no port", "example.com", "", "example.com", true),
		newJoinHostPortCase("subdomain with port", "sub.example.com", "80", "sub.example.com:80", true),

		// Invalid cases
		newJoinHostPortCase("empty host", "", "8080", "", false),
		newJoinHostPortCase("invalid hostname", "invalid host", "8080", "", false),
		newJoinHostPortCase("port 0 valid", "example.com", "0", "example.com:0", true),
		newJoinHostPortCase("port out of range", "example.com", "99999", "", false),
		newJoinHostPortCase("invalid port", "example.com", "invalid", "", false),
		newJoinHostPortCase("negative port", "example.com", "-1", "", false),
		newJoinHostPortCase("bad hostname dots", "bad..name", "8080", "", false),
		newJoinHostPortCase("bad hostname leading dot", ".invalid", "8080", "", false),
	)
}

func TestJoinHostPort(t *testing.T) {
	RunTestCases(t, joinHostPortTestCases())
}

type doMakeHostPortCase struct {
	name        string
	host        string
	port        string
	expected    string
	ok          bool
	defaultPort uint16
}

func (tc doMakeHostPortCase) Name() string {
	return tc.name
}

func (tc doMakeHostPortCase) Test(t *testing.T) {
	t.Helper()

	result, err := doMakeHostPort(tc.host, tc.port, tc.defaultPort)
	if (err == nil) != tc.ok || (tc.ok && result != tc.expected) {
		t.Errorf("doMakeHostPort(%q, %q, %d) -> %q, %v; expected %q, ok=%v",
			tc.host, tc.port, tc.defaultPort, result, err, tc.expected, tc.ok)
	} else if tc.ok {
		t.Logf("doMakeHostPort(%q, %q, %d) -> %q ✓", tc.host, tc.port, tc.defaultPort, result)
	} else {
		t.Logf("doMakeHostPort(%q, %q, %d) -> error: %v ✓", tc.host, tc.port, tc.defaultPort, err)
	}
}

//revive:disable-next-line:argument-limit
func newDoMakeHostPortCase(name, host, port string, defaultPort uint16, expected string, ok bool) doMakeHostPortCase {
	return doMakeHostPortCase{
		name:        name,
		host:        host,
		port:        port,
		expected:    expected,
		ok:          ok,
		defaultPort: defaultPort,
	}
}

func doMakeHostPortTestCases() []doMakeHostPortCase {
	return S(
		// Valid cases with explicit port
		newDoMakeHostPortCase("explicit port used", "example.com", "8080", 9000, "example.com:8080", true),
		newDoMakeHostPortCase("explicit port IPv4", "192.168.1.1", "443", 80, "192.168.1.1:443", true),
		newDoMakeHostPortCase("explicit port IPv6", "[::1]", "9000", 8080, "[::1]:9000", true),

		// Valid cases with default port
		newDoMakeHostPortCase("default port used", "example.com", "", 8080, "example.com:8080", true),
		newDoMakeHostPortCase("default port IPv4", "192.168.1.1", "", 443, "192.168.1.1:443", true),
		newDoMakeHostPortCase("default port IPv6", "[::1]", "", 9000, "[::1]:9000", true),

		// Valid cases with no port
		newDoMakeHostPortCase("no port hostname", "example.com", "", 0, "example.com", true),
		newDoMakeHostPortCase("no port IPv4", "192.168.1.1", "", 0, "192.168.1.1", true),
		newDoMakeHostPortCase("no port IPv6", "[::1]", "", 0, "[::1]", true),

		// Invalid cases
		newDoMakeHostPortCase("port 0 not allowed", "example.com", "0", 8080, "", false),
		newDoMakeHostPortCase("port 0 not allowed IPv4", "192.168.1.1", "0", 443, "", false),
		newDoMakeHostPortCase("port 0 not allowed IPv6", "[::1]", "0", 9000, "", false),
	)
}

func TestDoMakeHostPort(t *testing.T) {
	RunTestCases(t, doMakeHostPortTestCases())
}

type doJoinHostPortCase struct {
	name     string
	host     string
	port     string
	expected string
	ok       bool
}

func (tc doJoinHostPortCase) Name() string {
	return tc.name
}

func (tc doJoinHostPortCase) Test(t *testing.T) {
	t.Helper()

	result, err := doJoinHostPort(tc.host, tc.port)
	if (err == nil) != tc.ok || (tc.ok && result != tc.expected) {
		t.Errorf("doJoinHostPort(%q, %q) -> %q, %v; expected %q, ok=%v",
			tc.host, tc.port, result, err, tc.expected, tc.ok)
	} else if tc.ok {
		t.Logf("doJoinHostPort(%q, %q) -> %q ✓", tc.host, tc.port, result)
	} else {
		t.Logf("doJoinHostPort(%q, %q) -> error: %v ✓", tc.host, tc.port, err)
	}
}

func newDoJoinHostPortCase(name, host, port, expected string, ok bool) doJoinHostPortCase {
	return doJoinHostPortCase{
		name:     name,
		host:     host,
		port:     port,
		expected: expected,
		ok:       ok,
	}
}

func doJoinHostPortTestCases() []doJoinHostPortCase {
	return S(
		// Valid cases
		newDoJoinHostPortCase("valid hostname", "example.com", "8080", "example.com:8080", true),
		newDoJoinHostPortCase("valid IPv4", "192.168.1.1", "443", "192.168.1.1:443", true),
		newDoJoinHostPortCase("valid IPv6", "[::1]", "9000", "[::1]:9000", true),
		newDoJoinHostPortCase("valid hostname SSH", "localhost", "22", "localhost:22", true),
		newDoJoinHostPortCase("valid subdomain", "sub.example.com", "80", "sub.example.com:80", true),

		// Invalid cases
		newDoJoinHostPortCase("port 0 valid", "example.com", "0", "example.com:0", true),
		newDoJoinHostPortCase("port out of range", "192.168.1.1", "99999", "", false),
		newDoJoinHostPortCase("invalid port", "[::1]", "invalid", "", false),
		newDoJoinHostPortCase("negative port", "localhost", "-1", "", false),
		newDoJoinHostPortCase("port out of range high", "example.com", "65536", "", false),
		newDoJoinHostPortCase("non-numeric port", "test.com", "abc", "", false),
	)
}

func TestDoJoinHostPort(t *testing.T) {
	RunTestCases(t, doJoinHostPortTestCases())
}

type ipForHostPortCase struct {
	name     string
	input    string
	expected string
}

func (tc ipForHostPortCase) Name() string {
	return tc.name
}

func (tc ipForHostPortCase) Test(t *testing.T) {
	t.Helper()

	addr, err := ParseAddr(tc.input)
	if err != nil {
		t.Errorf("ParseAddr(%q) failed: %v", tc.input, err)
		return
	}

	result := ipForHostPort(addr)
	if result != tc.expected {
		t.Errorf("ipForHostPort(%q) -> %q; expected %q", tc.input, result, tc.expected)
	} else {
		t.Logf("ipForHostPort(%q) -> %q ✓", tc.input, result)
	}
}

func newIPForHostPortCase(name, input, expected string) ipForHostPortCase {
	return ipForHostPortCase{
		name:     name,
		input:    input,
		expected: expected,
	}
}

func ipForHostPortTestCases() []ipForHostPortCase {
	return S(
		// IPv4 addresses (should not be bracketed)
		newIPForHostPortCase("IPv4 localhost", "127.0.0.1", "127.0.0.1"),
		newIPForHostPortCase("IPv4 private", "192.168.1.1", "192.168.1.1"),
		newIPForHostPortCase("IPv4 private 10", "10.0.0.1", "10.0.0.1"),
		newIPForHostPortCase("IPv4 private 172", "172.16.0.1", "172.16.0.1"),
		newIPForHostPortCase("IPv4 unspecified", "0.0.0.0", "0.0.0.0"),
		newIPForHostPortCase("IPv4 broadcast", "255.255.255.255", "255.255.255.255"),

		// IPv6 addresses (should be bracketed)
		newIPForHostPortCase("IPv6 localhost", "::1", "[::1]"),
		newIPForHostPortCase("IPv6 unspecified", "::", "[::]"),
		newIPForHostPortCase("IPv6 example", "2001:db8::1", "[2001:db8::1]"),
		newIPForHostPortCase("IPv6 link-local", "fe80::1", "[fe80::1]"),
		newIPForHostPortCase("IPv6 full", "2001:db8:85a3::8a2e:370:7334", "[2001:db8:85a3::8a2e:370:7334]"),
		newIPForHostPortCase("IPv6 mapped", "::ffff:192.0.2.1", "[::ffff:192.0.2.1]"),
	)
}

func TestIPForHostPort(t *testing.T) {
	RunTestCases(t, ipForHostPortTestCases())
}

type addrErrCase struct {
	name string
	addr string
	why  string
}

func (tc addrErrCase) Name() string {
	return tc.name
}

func (tc addrErrCase) Test(t *testing.T) {
	t.Helper()

	err := addrErr(tc.addr, tc.why)

	// Check that it returns a *net.AddrError
	addrErr, ok := err.(*net.AddrError)
	if !ok {
		t.Errorf("addrErr() should return *net.AddrError, got %T", err)
		return
	}

	// Check the error details
	if addrErr.Addr != tc.addr {
		t.Errorf("addrErr.Addr = %q; expected %q", addrErr.Addr, tc.addr)
	}
	if addrErr.Err != tc.why {
		t.Errorf("addrErr.Err = %q; expected %q", addrErr.Err, tc.why)
	}

	// Check the error message format (net.AddrError formats as "address <addr>: <err>")
	var expectedMsg string
	if tc.addr == "" {
		expectedMsg = tc.why
	} else {
		expectedMsg = "address " + tc.addr + ": " + tc.why
	}
	if addrErr.Error() != expectedMsg {
		t.Errorf("addrErr.Error() = %q; expected %q", addrErr.Error(), expectedMsg)
	}

	t.Logf("addrErr(%q, %q) -> %v ✓", tc.addr, tc.why, err)
}

func newAddrErrCase(name, addr, why string) addrErrCase {
	return addrErrCase{
		name: name,
		addr: addr,
		why:  why,
	}
}

func addrErrTestCases() []addrErrCase {
	return S(
		newAddrErrCase("basic error", "invalid.address", "test error"),
		newAddrErrCase("empty addr", "", "empty address"),
		newAddrErrCase("empty reason", "example.com", ""),
		newAddrErrCase("special chars", "test@example.com", "invalid format"),
	)
}

func TestAddrErr(t *testing.T) {
	RunTestCases(t, addrErrTestCases())
}
