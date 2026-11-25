package core

import (
	"net"
	"strings"
	"testing"
	"unicode/utf8"
)

var (
	_ TestCase = parseHostPortTestCase{}
	_ TestCase = countRunesFuncTestCase{}
)

type parseHostPortTestCase struct {
	name     string
	hostPort string
	host     string
	zone     string
	port     string
	rejected bool
}

func (tc parseHostPortTestCase) Name() string {
	return tc.name
}

func (tc parseHostPortTestCase) Test(t *testing.T) {
	host, zone, port, err := parseHostPort(tc.hostPort)

	AssertEqual(t, tc.host, host, "host")
	AssertEqual(t, tc.zone, zone, "zone")
	AssertEqual(t, tc.port, port, "port")

	if tc.rejected {
		// A rejection names the whole input.
		rejection := AssertMustErrorAs[*net.AddrError](t, err, "error")
		AssertEqual(t, tc.hostPort, rejection.Addr, "error address")
		return
	}

	AssertNoError(t, err, "error")
}

// newParseHostPortTestCase declares an accepted input by the
// host, zone and port it splits into, each empty where the input
// carries none.
func newParseHostPortTestCase(name, hostPort, host, zone, port string) parseHostPortTestCase {
	return parseHostPortTestCase{
		name:     name,
		hostPort: MustNotZeroForTest(hostPort),
		host:     host,
		zone:     zone,
		port:     port,
	}
}

// newParseHostPortTestCaseRejected declares an input
// parseHostPort rejects, returning empty strings.
func newParseHostPortTestCaseRejected(name, hostPort string) parseHostPortTestCase {
	return parseHostPortTestCase{
		name:     name,
		hostPort: hostPort,
		rejected: true,
	}
}

func parseHostPortTestCases() []parseHostPortTestCase {
	return S(
		// A zone splits off as written, case included, so that the
		// mapping the host goes through later cannot fold it. The
		// address before it is left as given too.
		newParseHostPortTestCase("bracketed zone", "[fe80::1%ETH0]:80", "fe80::1", "ETH0", "80"),
		newParseHostPortTestCase("bracketed zone no port", "[fe80::1%ETH0]", "fe80::1", "ETH0", ""),
		newParseHostPortTestCase("zone without brackets", "fe80::1%ETH0", "fe80::1", "ETH0", ""),
		newParseHostPortTestCase("full-width address zone", "[fe80::１%ETH0]:80", "fe80::１", "ETH0", "80"),

		// The runes UTS 46 maps into '%' separate a zone as '%' does.
		newParseHostPortTestCase("full-width percent zone", "[fe80::1％ETH0]:80", "fe80::1", "ETH0", "80"),
		newParseHostPortTestCase("small percent zone", "[fe80::1﹪ETH0]:80", "fe80::1", "ETH0", "80"),

		// The parser checks no character of a zone but a bracket, a ':',
		// a separator and NUL: in brackets it runs to the first ']', and
		// without them to the end, or to the ':' where the input holds
		// no other.
		newParseHostPortTestCase("zone with a space", "[fe80::1%Ethernet 2]:80", "fe80::1", "Ethernet 2", "80"),
		newParseHostPortTestCase("zone after a name", "a%b:80", "a", "b", "80"),
		newParseHostPortTestCase("zone after a name no port", "a%b.name", "a", "b.name", ""),
		// The bracket closes at the first ']', leaving the rest to the
		// port, whose check refuses it.
		newParseHostPortTestCase("zone before a second ']'", "[fe80::1%x]:80]:9", "fe80::1", "x", "80]:9"),

		// Without a zone, a host splits from its port, brackets removed,
		// and neither is checked.
		newParseHostPortTestCase("name", "name", "name", "", ""),
		newParseHostPortTestCase("name and port", "name:80", "name", "", "80"),
		newParseHostPortTestCase("name and word port", "name:port", "name", "", "port"),
		newParseHostPortTestCase("port only", ":80", "", "", "80"),
		newParseHostPortTestCase("bracketed IPv6 and port", "[::1]:80", "::1", "", "80"),
		newParseHostPortTestCase("bracketed name and port", "[name]:80", "name", "", "80"),
		// An IPv6 address without brackets is read whole, its last group
		// included, whichever of the runes UTS 46 maps into ':' it is
		// spelt with.
		newParseHostPortTestCase("IPv6 trailing group", "::1:80", "::1:80", "", ""),
		newParseHostPortTestCase("full-width IPv6 trailing group", "：：１:80", "：：１:80", "", ""),

		// A separator needs an address before it and a zone after it.
		newParseHostPortTestCaseRejected("zone alone", "%eth0"),
		newParseHostPortTestCaseRejected("zone alone in brackets", "[%eth0]"),
		newParseHostPortTestCaseRejected("empty zone", "fe80::1%"),
		newParseHostPortTestCaseRejected("empty zone before port", "a%:80"),
		newParseHostPortTestCaseRejected("empty bracketed zone", "[fe80::1%]"),
		newParseHostPortTestCaseRejected("empty full-width percent zone", "fe80::1％"),
		newParseHostPortTestCaseRejected("full-width percent zone alone", "％eth0"),

		// A zone holds no bracket, which would leave an address joined
		// in brackets that does not split back, no ':', which no Linux
		// interface name holds, no separator, and no NUL.
		newParseHostPortTestCaseRejected("zone with a ']'", "fe80::1%a]"),
		newParseHostPortTestCaseRejected("zone with a ']' before port", "a%b]:80"),
		newParseHostPortTestCaseRejected("zone with a '['", "fe80::1%a[b"),
		newParseHostPortTestCaseRejected("bracketed zone with a '['", "[fe80::1%a[b]:80"),
		newParseHostPortTestCaseRejected("zone with a ':'", "fe80::1%a:b"),
		newParseHostPortTestCaseRejected("bracketed zone with a ':'", "[fe80::1%a:b]:80"),
		newParseHostPortTestCaseRejected("zone with a separator", "[fe80::1%a%b]:80"),
		newParseHostPortTestCaseRejected("zone with a full-width percent", "fe80::1%a％b"),
		newParseHostPortTestCaseRejected("zone with a small percent", "[fe80::1%a﹪b]"),
		newParseHostPortTestCaseRejected("zone with a NUL", "[fe80::1%a\x00b]:80"),

		// A ':' carries a port, and brackets close, followed by nothing
		// or a ':' and a port.
		newParseHostPortTestCaseRejected("name empty port", "name:"),
		newParseHostPortTestCaseRejected("bracketed empty port", "[::1]:"),
		newParseHostPortTestCaseRejected("unclosed brackets", "[::1"),
		newParseHostPortTestCaseRejected("unclosed bracketed zone", "[fe80::1%eth0"),
		newParseHostPortTestCaseRejected("empty brackets", "[]"),
		newParseHostPortTestCaseRejected("bracketed trailing character", "[::1]x"),

		// The split rejects an empty input.
		newParseHostPortTestCaseRejected("empty", ""),
	)
}

func TestParseHostPort(t *testing.T) {
	RunTestCases(t, parseHostPortTestCases())
}

// TestZoneSeparators states that the runes nameProfile maps into a '%'
// are '%', '﹪' and '％', and that zoneSeparatorLen takes each of them
// as a separator, so that a change in the mapping turning another rune
// into a '%', which would fold the zone after it, fails here.
func TestZoneSeparators(t *testing.T) {
	var mapped []rune
	for r := rune(0); r <= utf8.MaxRune; r++ {
		if !utf8.ValidRune(r) {
			// surrogate, no rune of a string
			continue
		}

		// ToASCII returns what it mapped even where it refuses the
		// result, and the mapping is what turns a rune into a '%'.
		s, _ := nameProfile.ToASCII(string(r))
		if strings.Contains(s, "%") {
			mapped = append(mapped, r)
		}
	}

	AssertSliceEqual(t, S('%', '﹪', '％'), mapped, "runes mapped to percent")
	for _, r := range mapped {
		AssertEqual(t, utf8.RuneLen(r), zoneSeparatorLen(string(r)), "separator %q", r)
	}
}

// TestColonRunes states that the runes nameProfile maps into a ':' are
// ':', '︓', '﹕' and '：', and that isColon takes each of them, so
// that a change in the mapping turning another rune into a ':', which
// would leave an IPv6 address without brackets read as a host and a
// port, fails here. '⩴' maps into "::=", whose '=' no host holds, so
// the splits refuse it as they read it.
func TestColonRunes(t *testing.T) {
	var mapped []rune
	for r := rune(0); r <= utf8.MaxRune; r++ {
		if !utf8.ValidRune(r) {
			// surrogate, no rune of a string
			continue
		}

		// ToASCII returns what it mapped even where it refuses the
		// result, and the mapping is what turns a rune into a ':'.
		s, _ := nameProfile.ToASCII(string(r))
		if s == ":" {
			mapped = append(mapped, r)
		}
	}

	AssertSliceEqual(t, S(':', '︓', '﹕', '：'), mapped, "runes mapped to colon")
	for _, r := range mapped {
		AssertTrue(t, isColon(r), "colon %q", r)
	}
}

type countRunesFuncTestCase struct {
	name     string
	s        string
	expected int
}

func (tc countRunesFuncTestCase) Name() string {
	return tc.name
}

func (tc countRunesFuncTestCase) Test(t *testing.T) {
	AssertEqual(t, tc.expected, countRunesFunc(tc.s, isColon), "count")
}

// newCountRunesFuncTestCase declares a string by the number of colons
// countRunesFunc counts in it through isColon.
func newCountRunesFuncTestCase(name, s string, expected int) countRunesFuncTestCase {
	return countRunesFuncTestCase{
		name:     name,
		s:        s,
		expected: expected,
	}
}

func countRunesFuncTestCases() []countRunesFuncTestCase {
	return S(
		// Every rune the predicate takes counts, whichever it is.
		newCountRunesFuncTestCase("ASCII colons", "::1", 2),
		newCountRunesFuncTestCase("mixed colons", "：：１:80", 3),

		// Nothing counts where the predicate takes no rune.
		newCountRunesFuncTestCase("no colon", "name", 0),
		newCountRunesFuncTestCase("empty", "", 0),
	)
}

func TestCountRunesFunc(t *testing.T) {
	RunTestCases(t, countRunesFuncTestCases())
}
