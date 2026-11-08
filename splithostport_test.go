package core_test

// cspell:words alef ｇｏｏｄ

import (
	"net"
	"net/netip"
	"strings"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = splitAddrPortTestCase{}
	_ core.TestCase = splitHostPortTestCase{}
	_ core.TestCase = splitHostPortASCIITestCase{}
	_ core.TestCase = makeHostPortTestCase{}
	_ core.TestCase = joinHostPortTestCase{}
)

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
	addr, port, err := core.SplitAddrPort(tc.addrPort)

	core.AssertEqual(t, tc.addr, addr, "address")
	core.AssertEqual(t, tc.port, port, "port")

	if tc.rejected {
		// A rejection names the whole input.
		rejection := core.AssertMustErrorAs[*net.AddrError](t, err, "error")
		core.AssertEqual(t, tc.addrPort, rejection.Addr, "error address")
		return
	}

	core.AssertNoError(t, err, "error")
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
	return core.S(
		// An IP address splits from its port, which is 0 where the
		// input carries none.
		newSplitAddrPortTestCase("unspecified IPv4", "0.0.0.0:6060", "0.0.0.0", 6060),
		newSplitAddrPortTestCase("IPv6 no port", "::1", "::1", 0),
		newSplitAddrPortTestCase("bracketed IPv6 no port", "[::1]", "::1", 0),
		newSplitAddrPortTestCase("bracketed IPv6 and port", "[::1]:1234", "::1", 1234),
		newSplitAddrPortTestCase("bracketed IPv6 and port 0", "[::1]:0", "::1", 0),
		newSplitAddrPortTestCase("IPv6 zone", "[fe80::1%eth0]:80", "fe80::1%eth0", 80),
		newSplitAddrPortTestCase("unspecified IPv6", "[::]:6060", "::", 6060),
		// A port alone gets the undetermined host.
		newSplitAddrPortTestCase("port only", ":6060", "::", 6060),

		// A host UTS 46 maps into an IP address comes back as that
		// address.
		newSplitAddrPortTestCase("ideographic full stops IPv4", "１２７。０。０。１:80", "127.0.0.1", 80),
		newSplitAddrPortTestCase("full-width IPv6", "：：１", "::1", 0),

		// A port that is not a number is rejected.
		newSplitAddrPortTestCaseRejected("bracketed IPv6 word port", "[::1]:port"),

		// A host that is not an IP address is rejected, a valid name
		// included.
		newSplitAddrPortTestCaseRejected("name and port", "name:1234"),
		newSplitAddrPortTestCaseRejected("invalid punycode", "xn--a:80"),

		// The split rejects an empty input.
		newSplitAddrPortTestCaseRejected("empty", ""),
	)
}

func TestSplitAddrPort(t *testing.T) {
	core.RunTestCases(t, splitAddrPortTestCases())
}

type splitHostPortTestCase struct {
	name     string
	hostPort string
	host     string
	port     string
	rejected bool
}

func (tc splitHostPortTestCase) Name() string {
	return tc.name
}

func (tc splitHostPortTestCase) Test(t *testing.T) {
	host, port, err := core.SplitHostPort(tc.hostPort)

	core.AssertEqual(t, tc.host, host, "host")
	core.AssertEqual(t, tc.port, port, "port")

	if tc.rejected {
		// A rejection names the whole input.
		rejection := core.AssertMustErrorAs[*net.AddrError](t, err, "error")
		core.AssertEqual(t, tc.hostPort, rejection.Addr, "error address")
		return
	}

	core.AssertMustNoError(t, err, "error")

	// JoinHostPort joins the host and port again, into a string that
	// splits back into the same pair.
	joined, err := core.JoinHostPort(host, port)
	core.AssertMustNoError(t, err, "join error")
	againHost, againPort, err := core.SplitHostPort(joined)
	core.AssertMustNoError(t, err, "resplit error")
	core.AssertEqual(t, host, againHost, "resplit host")
	core.AssertEqual(t, port, againPort, "resplit port")
}

// newSplitHostPortTestCase declares an accepted input by the host and
// port it splits into; the port is empty where the input carries none.
func newSplitHostPortTestCase(name, hostPort, host, port string) splitHostPortTestCase {
	return splitHostPortTestCase{
		name:     name,
		hostPort: hostPort,
		host:     core.MustNotZeroForTest(host),
		port:     port,
	}
}

// newSplitHostPortTestCaseRejected declares an input SplitHostPort
// rejects, returning empty strings.
func newSplitHostPortTestCaseRejected(name, hostPort string) splitHostPortTestCase {
	return splitHostPortTestCase{
		name:     name,
		hostPort: hostPort,
		rejected: true,
	}
}

func splitHostPortTestCases() []splitHostPortTestCase {
	label63 := strings.Repeat("a", 63)
	label64 := label63 + "a"
	name253 := strings.Join(core.S(label63, label63, label63, label63[2:]), ".")
	name254 := name253 + "a"
	label66UTF8 := strings.Repeat("世", 22)

	return core.S(
		// A name splits from its port, which comes back in canonical
		// form.
		newSplitHostPortTestCase("name", "name", "name", ""),
		newSplitHostPortTestCase("dotted name", "good.name", "good.name", ""),
		newSplitHostPortTestCase("name and port", "name:1234", "name", "1234"),
		newSplitHostPortTestCase("name and port 0", "name:0", "name", "0"),
		newSplitHostPortTestCase("name and port 65535", "name:65535", "name", "65535"),
		newSplitHostPortTestCase("name and padded port", "name:0080", "name", "80"),

		// A name comes back as UTS 46 maps it, in Unicode.
		newSplitHostPortTestCase("international name", "Hello.世界", "hello.世界", ""),
		newSplitHostPortTestCase("full-width name", "ｇｏｏｄ.name", "good.name", ""),
		newSplitHostPortTestCase("ideographic full stop", "good。name", "good.name", ""),
		newSplitHostPortTestCase("punycode name", "hello.xn--rhqv96g", "hello.世界", ""),
		newSplitHostPortTestCase("punycode middle dot", "xn--ll-0ea", "l·l", ""),
		newSplitHostPortTestCase("ignored rune", "a\u200Bb.name", "ab.name", ""),
		newSplitHostPortTestCase("symbol", "☃.name", "☃.name", ""),

		// A label holds letters, digits, '_' and '-', and only the last
		// label of a name is judged as a number.
		newSplitHostPortTestCase("underscore label", "a_b.name", "a_b.name", ""),
		newSplitHostPortTestCase("double hyphen label", "ab--c.name", "ab--c.name", ""),
		// A hyphen at an end of a label's ASCII characters is not at an
		// end of the label when a character outside ASCII lies beyond it.
		newSplitHostPortTestCase("hyphen after international character", "世-a.name", "世-a.name", ""),
		newSplitHostPortTestCase("hyphen before international character", "a-世.name", "a-世.name", ""),
		newSplitHostPortTestCase("numeric label", "123.name", "123.name", ""),
		newSplitHostPortTestCase("non-hexadecimal 0x label", "name.0xg", "name.0xg", ""),

		// A label holds up to 63 octets, and a name up to 253, counted
		// in the ASCII form.
		newSplitHostPortTestCase("63-octet label", label63, label63, ""),
		newSplitHostPortTestCase("253-octet name", name253, name253, ""),
		newSplitHostPortTestCase("66-octet UTF-8 label", label66UTF8, label66UTF8, ""),

		// A name UTS 46 maps into an IP address comes back as that
		// address.
		newSplitHostPortTestCase("full-width zero", "０", "0.0.0.0", ""),
		newSplitHostPortTestCase("ideographic full stops IPv4", "１２７。０。０。１", "127.0.0.1", ""),
		newSplitHostPortTestCase("full-width IPv6", "：：１", "::1", ""),

		// An IP address splits from its port and comes back in
		// canonical text.
		newSplitHostPortTestCase("unspecified IPv4", "0.0.0.0:6060", "0.0.0.0", "6060"),
		// "0" is read as an IP address before it can be taken as a name.
		newSplitHostPortTestCase("unspecified IPv4 short", "0:6060", "0.0.0.0", "6060"),
		newSplitHostPortTestCase("IPv6 no port", "::1", "::1", ""),
		// An unbracketed IPv6 address is read whole, its last group
		// included, so a port needs the brackets.
		newSplitHostPortTestCase("IPv6 trailing group", "::1:8080", "::1:8080", ""),
		newSplitHostPortTestCase("bracketed IPv6 no port", "[::1]", "::1", ""),
		newSplitHostPortTestCase("bracketed IPv6 and port", "[::1]:1234", "::1", "1234"),
		newSplitHostPortTestCase("non-canonical IPv6", "[2001:DB8:0::1]:80", "2001:db8::1", "80"),
		newSplitHostPortTestCase("IPv6 zone", "[fe80::1%eth0]:80", "fe80::1%eth0", "80"),
		newSplitHostPortTestCase("unspecified IPv6", "[::]:6060", "::", "6060"),
		// A port alone gets the undetermined host.
		newSplitHostPortTestCase("port only", ":6060", "::", "6060"),

		// Brackets come off any host, a name or an IPv4 address as well
		// as an IPv6 one.
		newSplitHostPortTestCase("bracketed name", "[name]:80", "name", "80"),
		newSplitHostPortTestCase("bracketed IPv4", "[192.0.2.1]:80", "192.0.2.1", "80"),

		// A ':' carries a port, a decimal number from 0 to 65535.
		newSplitHostPortTestCaseRejected("name empty port", "name:"),
		newSplitHostPortTestCaseRejected("name decimal-point port", "name:123.4"),
		newSplitHostPortTestCaseRejected("name negative port", "name:-1"),
		newSplitHostPortTestCaseRejected("name port out of range", "name:65536"),
		newSplitHostPortTestCaseRejected("name non-numeric port", "name:port"),
		newSplitHostPortTestCaseRejected("bracketed IPv6 empty port", "[::1]:"),

		// A host that is not an IP address is a name, and is rejected
		// unless it is a valid one.
		newSplitHostPortTestCaseRejected("empty label", "bad..name"),
		newSplitHostPortTestCaseRejected("leading dot", ".name"),
		newSplitHostPortTestCaseRejected("trailing dot", "name."),
		// idna drops a soft hyphen or a combining grapheme joiner, and
		// the label it leaves empty is rejected.
		newSplitHostPortTestCaseRejected("soft hyphen label", "bad.\u00AD.name"),
		newSplitHostPortTestCaseRejected("combining grapheme joiner label", "bad.\u034F.name"),
		// idna refuses these, the alef symbol only once the mapping has
		// made it a Hebrew letter and the name is decoded.
		newSplitHostPortTestCaseRejected("invalid punycode", "xn--a.name"),
		newSplitHostPortTestCaseRejected("alef symbol label", "aℵ.name"),
		newSplitHostPortTestCaseRejected("mixed direction label", "\u05D0a.name"),
		// A label holds no other character, nor a hyphen at either end.
		newSplitHostPortTestCaseRejected("space", "bad name"),
		newSplitHostPortTestCaseRejected("percent label", "a%b.name"),
		newSplitHostPortTestCaseRejected("plus label", "a+b.name"),
		newSplitHostPortTestCaseRejected("leading hyphen", "-name"),
		newSplitHostPortTestCaseRejected("trailing hyphen", "name-"),
		// The hyphen is judged in the label as decoded, as its ASCII
		// form, xn----9o6a for -世, does not show it.
		newSplitHostPortTestCaseRejected("leading hyphen international label", "-世.name"),
		newSplitHostPortTestCaseRejected("trailing hyphen international label", "世-.name"),
		newSplitHostPortTestCaseRejected("leading hyphen punycode label", "xn----9o6a.name"),
		// A label is rejected beyond 63 octets, and a name beyond 253.
		newSplitHostPortTestCaseRejected("64-octet label", label64),
		newSplitHostPortTestCaseRejected("254-octet name", name254),
		// A name whose last label is a number, all digits or "0x" and
		// hexadecimal digits, is rejected unless it is an IP address.
		newSplitHostPortTestCaseRejected("numeric name", "1234"),
		newSplitHostPortTestCaseRejected("out-of-range IPv4", "256.1.1.1"),
		newSplitHostPortTestCaseRejected("numeric last label", "name.1"),
		newSplitHostPortTestCaseRejected("hexadecimal last label", "name.0x1f"),
		newSplitHostPortTestCaseRejected("bare 0x last label", "name.0x"),
		newSplitHostPortTestCaseRejected("hexadecimal IPv4", "0x7f.1"),
		newSplitHostPortTestCaseRejected("full-width hexadecimal IPv4", "０ｘ７ｆ。１"),
		// An empty bracket pair holds an empty host, which is rejected,
		// where ":port" alone gets the undetermined host.
		newSplitHostPortTestCaseRejected("empty brackets", "[]"),

		// Brackets close, followed by nothing or a ':' and a port.
		newSplitHostPortTestCaseRejected("incomplete bracketed IPv6", "[::1:1234"),
		newSplitHostPortTestCaseRejected("bracketed IPv6 trailing character", "[::1]x"),

		// The split rejects an empty input.
		newSplitHostPortTestCaseRejected("empty", ""),
	)
}

func TestSplitHostPort(t *testing.T) {
	core.RunTestCases(t, splitHostPortTestCases())
}

type splitHostPortASCIITestCase struct {
	name     string
	hostPort string
	host     string
	port     string
	rejected bool
}

func (tc splitHostPortASCIITestCase) Name() string {
	return tc.name
}

func (tc splitHostPortASCIITestCase) Test(t *testing.T) {
	host, port, err := core.SplitHostPortASCII(tc.hostPort)

	core.AssertEqual(t, tc.host, host, "host")
	core.AssertEqual(t, tc.port, port, "port")

	if tc.rejected {
		// A rejection names the whole input, which SplitHostPort
		// rejects as well.
		rejection := core.AssertMustErrorAs[*net.AddrError](t, err, "error")
		core.AssertEqual(t, tc.hostPort, rejection.Addr, "error address")
		_, _, err = core.SplitHostPort(tc.hostPort)
		core.AssertError(t, err, "unicode error")
		return
	}

	core.AssertMustNoError(t, err, "error")

	// The ASCII host splits back unchanged, and SplitHostPort decodes it
	// into the host it gives for the input.
	again, _, err := core.SplitHostPortASCII(host)
	core.AssertMustNoError(t, err, "resplit error")
	core.AssertEqual(t, host, again, "resplit host")
	want, _, err := core.SplitHostPort(tc.hostPort)
	core.AssertMustNoError(t, err, "unicode error")
	decoded, _, err := core.SplitHostPort(host)
	core.AssertMustNoError(t, err, "decode error")
	core.AssertEqual(t, want, decoded, "decoded host")
}

// newSplitHostPortASCIITestCase declares an accepted input by the host
// and port it splits into; the port is empty where the input carries
// none.
func newSplitHostPortASCIITestCase(name, hostPort, host, port string) splitHostPortASCIITestCase {
	return splitHostPortASCIITestCase{
		name:     name,
		hostPort: hostPort,
		host:     core.MustNotZeroForTest(host),
		port:     port,
	}
}

// newSplitHostPortASCIITestCaseRejected declares an input
// SplitHostPortASCII rejects, returning empty strings.
func newSplitHostPortASCIITestCaseRejected(name, hostPort string) splitHostPortASCIITestCase {
	return splitHostPortASCIITestCase{
		name:     name,
		hostPort: hostPort,
		rejected: true,
	}
}

func splitHostPortASCIITestCases() []splitHostPortASCIITestCase {
	return core.S(
		// A name comes back as UTS 46 maps it, in ASCII, the port in
		// canonical form.
		newSplitHostPortASCIITestCase("name and padded port", "Name:0080", "name", "80"),
		newSplitHostPortASCIITestCase("international name", "Hello.世界", "hello.xn--rhqv96g", ""),
		newSplitHostPortASCIITestCase("punycode name", "hello.xn--rhqv96g", "hello.xn--rhqv96g", ""),
		newSplitHostPortASCIITestCase("full-width name", "ｇｏｏｄ.name", "good.name", ""),
		newSplitHostPortASCIITestCase("middle dot", "l·l", "xn--ll-0ea", ""),

		// A label holds letters, digits, '_' and '-'.
		newSplitHostPortASCIITestCase("underscore label", "a_b.name", "a_b.name", ""),
		newSplitHostPortASCIITestCase("double hyphen label", "ab--c.name", "ab--c.name", ""),

		// An IP address comes back in canonical text, a host UTS 46 maps
		// into one included.
		newSplitHostPortASCIITestCase("non-canonical IPv6", "[2001:DB8:0::1]:80", "2001:db8::1", "80"),
		newSplitHostPortASCIITestCase("ideographic full stops IPv4", "１２７。０。０。１:80", "127.0.0.1", "80"),
		newSplitHostPortASCIITestCase("port only", ":6060", "::", "6060"),

		// What SplitHostPort rejects before decoding a name,
		// SplitHostPortASCII rejects.
		newSplitHostPortASCIITestCaseRejected("invalid punycode", "xn--a.name"),
		newSplitHostPortASCIITestCaseRejected("leading hyphen punycode label", "xn----9o6a.name"),
		newSplitHostPortASCIITestCaseRejected("numeric name", "1234"),
		newSplitHostPortASCIITestCaseRejected("name non-numeric port", "name:port"),
		newSplitHostPortASCIITestCaseRejected("empty", ""),
	)
}

func TestSplitHostPortASCII(t *testing.T) {
	core.RunTestCases(t, splitHostPortASCIITestCases())
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
	got, err := core.MakeHostPort(tc.hostPort, tc.defaultPort)

	core.AssertEqual(t, tc.expected, got, "host:port")

	if tc.rejected {
		// A rejection names the whole input.
		rejection := core.AssertMustErrorAs[*net.AddrError](t, err, "error")
		core.AssertEqual(t, tc.hostPort, rejection.Addr, "error address")
		return
	}

	core.AssertMustNoError(t, err, "error")

	// The other side splits the result with SplitHostPort, and the host
	// and port it returns join again into the same string.
	host, port, err := core.SplitHostPort(got)
	core.AssertMustNoError(t, err, "split error")
	again, err := core.JoinHostPort(host, port)
	core.AssertMustNoError(t, err, "rejoin error")
	core.AssertEqual(t, got, again, "rejoined host:port")
}

// newMakeHostPortTestCase declares an accepted input by the host:port
// it produces.
func newMakeHostPortTestCase(name, hostPort string, defaultPort uint16, expected string) makeHostPortTestCase {
	return makeHostPortTestCase{
		name:        name,
		hostPort:    hostPort,
		expected:    core.MustNotZeroForTest(expected),
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
	return core.S(
		// An IP address comes back in canonical text, an IPv6 one
		// bracketed when a port follows it.
		newMakeHostPortTestCase("unspecified IPv4 short", "0", 80, "0.0.0.0:80"),
		newMakeHostPortTestCase("IPv6 bracketed no port", "[::1]", 0, "::1"),
		newMakeHostPortTestCase("IPv6 unbracketed default port", "::1", 8080, "[::1]:8080"),
		newMakeHostPortTestCase("IPv6 zone", "[fe80::1%eth0]:80", 0, "[fe80::1%eth0]:80"),
		// A port alone gets the undetermined host.
		newMakeHostPortTestCase("port only", ":6060", 0, "[::]:6060"),

		// A name comes back cleaned, with its own port, else the
		// default one, else portless.
		newMakeHostPortTestCase("name explicit port", "example.com:80", 443, "example.com:80"),
		newMakeHostPortTestCase("name padded port", "example.com:0080", 443, "example.com:80"),
		newMakeHostPortTestCase("name mixed case", "Example.com:80", 443, "example.com:80"),
		newMakeHostPortTestCase("punycode name", "hello.xn--rhqv96g:80", 443, "hello.世界:80"),
		newMakeHostPortTestCase("name default port", "example.com", 443, "example.com:443"),
		newMakeHostPortTestCase("name no port", "example.com", 0, "example.com"),

		// A name UTS 46 maps into an IP address comes back as that
		// address.
		newMakeHostPortTestCase("ideographic full stops IPv4", "１２７。０。０。１", 80, "127.0.0.1:80"),

		// What SplitHostPort rejects, MakeHostPort rejects.
		newMakeHostPortTestCaseRejected("empty input", "", 8080),
		newMakeHostPortTestCaseRejected("space", "invalid host", 8080),
		newMakeHostPortTestCaseRejected("word port", "example.com:invalid", 8080),

		// Port 0 in the input is rejected, in any spelling.
		newMakeHostPortTestCaseRejected("port 0", "example.com:0", 8080),
		newMakeHostPortTestCaseRejected("port 0 padded", "example.com:00", 8080),
		// With no default, port 0 is still rejected, where
		// "example.com" alone comes back portless.
		newMakeHostPortTestCaseRejected("port 0 without default", "example.com:0", 0),
		// Port 0 is judged after the split has cleaned the host and the
		// port, and the error names the input.
		newMakeHostPortTestCaseRejected("port 0 cleaned name", "Example.com:0", 8080),
		newMakeHostPortTestCaseRejected("port 0 no host", ":0", 8080),
	)
}

func TestMakeHostPort(t *testing.T) {
	core.RunTestCases(t, makeHostPortTestCases())
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
	got, err := core.JoinHostPort(tc.host, tc.port)

	core.AssertEqual(t, tc.expected, got, "host:port")

	if tc.rejected {
		rejection := core.AssertMustErrorAs[*net.AddrError](t, err, "error")
		core.AssertEqual(t, tc.errAddr, rejection.Addr, "error address")
		return
	}

	core.AssertMustNoError(t, err, "error")

	// SplitHostPort splits the result back into its host and port,
	// which join again into the same string.
	host, port, err := core.SplitHostPort(got)
	core.AssertMustNoError(t, err, "split error")
	again, err := core.JoinHostPort(host, port)
	core.AssertMustNoError(t, err, "rejoin error")
	core.AssertEqual(t, got, again, "rejoined host:port")
}

// newJoinHostPortTestCase declares an accepted pair by the host:port
// it joins into.
func newJoinHostPortTestCase(name, host, port, expected string) joinHostPortTestCase {
	return joinHostPortTestCase{
		name:     name,
		host:     host,
		port:     port,
		expected: core.MustNotZeroForTest(expected),
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
	return core.S(
		// An IP address joins in canonical text, an IPv6 one bracketed
		// when a port follows it.
		newJoinHostPortTestCase("IPv6 with port", "::1", "8080", "[::1]:8080"),
		newJoinHostPortTestCase("IPv6 no port", "::1", "", "::1"),
		newJoinHostPortTestCase("non-canonical IPv6 with port", "2001:DB8:0::1", "80", "[2001:db8::1]:80"),
		newJoinHostPortTestCase("non-canonical IPv6 no port", "2001:DB8:0::1", "", "2001:db8::1"),
		newJoinHostPortTestCase("IPv6 zone", "fe80::1%eth0", "80", "[fe80::1%eth0]:80"),

		// A name joins cleaned, the port in canonical form.
		newJoinHostPortTestCase("name with port", "example.com", "443", "example.com:443"),
		newJoinHostPortTestCase("name no port", "example.com", "", "example.com"),
		newJoinHostPortTestCase("name mixed case with port", "Example.com", "443", "example.com:443"),
		newJoinHostPortTestCase("name mixed case no port", "Example.com", "", "example.com"),
		newJoinHostPortTestCase("punycode name", "hello.xn--rhqv96g", "443", "hello.世界:443"),
		newJoinHostPortTestCase("padded port", "example.com", "0080", "example.com:80"),

		// A name UTS 46 maps into an IP address joins as that address,
		// an IPv6 one bracketed when a port follows it.
		newJoinHostPortTestCase("full-width IPv6 with port", "：：１", "80", "[::1]:80"),

		// Port 0 joins, where MakeHostPort rejects it in its input.
		newJoinHostPortTestCase("port 0", "example.com", "0", "example.com:0"),

		// A rejected host is named alone, before the port is read.
		newJoinHostPortTestCaseRejected("empty host", "", "8080", ""),
		newJoinHostPortTestCaseRejected("space", "invalid host", "8080", "invalid host"),
		newJoinHostPortTestCaseRejected("space and word port", "invalid host", "invalid", "invalid host"),
		// Brackets belong to a host:port string, and a host carrying
		// them fails as both an IP and a name.
		newJoinHostPortTestCaseRejected("bracketed IPv6 host", "[::1]", "8080", "[::1]"),

		// A rejected port is named with the host in its cleaned form.
		newJoinHostPortTestCaseRejected("negative port", "example.com", "-1", "example.com:-1"),
		newJoinHostPortTestCaseRejected("port 65536", "example.com", "65536", "example.com:65536"),
		newJoinHostPortTestCaseRejected("word port mixed case", "Example.com", "invalid", "example.com:invalid"),
		newJoinHostPortTestCaseRejected("word port IPv6", "::1", "invalid", "[::1]:invalid"),
	)
}

func TestJoinHostPort(t *testing.T) {
	core.RunTestCases(t, joinHostPortTestCases())
}
