package core

import (
	"net"
	"net/netip"
	"testing"
)

var _ TestCase = typeSpecificAddrPortTestCase{}

// Test typeSpecificAddrPort directly
type typeSpecificAddrPortTestCase struct {
	name      string
	input     any
	want      netip.AddrPort
	wantOK    bool
	wantKnown bool
}

var typeSpecificAddrPortTestCases = []typeSpecificAddrPortTestCase{
	newTypeSpecificAddrPortTestCase("AddrPort value", netip.MustParseAddrPort("192.168.1.1:8080"),
		netip.MustParseAddrPort("192.168.1.1:8080"), true),
	newTypeSpecificAddrPortTestCase("AddrPort pointer", NewAddrPortPtrForTest(netip.MustParseAddrPort("10.0.0.1:443")),
		netip.MustParseAddrPort("10.0.0.1:443"), true),
	newTypeSpecificAddrPortTestCase("TCPAddr", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9000},
		netip.MustParseAddrPort("127.0.0.1:9000"), true),
	newTypeSpecificAddrPortTestCase("UDPAddr", &net.UDPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 53},
		netip.MustParseAddrPort("10.0.0.1:53"), true),
	newTypeSpecificAddrPortTestCase("zero AddrPort", netip.AddrPort{}, netip.AddrPort{}, false),
	newTypeSpecificAddrPortTestCase("pointer to zero AddrPort", NewAddrPortPtrForTest(netip.AddrPort{}),
		netip.AddrPort{}, false),
	newTypeSpecificAddrPortTestCase("TCPAddr with invalid IP", &net.TCPAddr{IP: S[byte](1, 2, 3), Port: 80},
		netip.AddrPort{}, false), // Invalid IP length
	newTypeSpecificAddrPortTestCase("UDPAddr with invalid IP", &net.UDPAddr{IP: S[byte](), Port: 80},
		netip.AddrPort{}, false), // Empty IP
	newTypeSpecificAddrPortTestCase("TCPAddr with negative port",
		&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: -1}, netip.AddrPort{}, false),
	newTypeSpecificAddrPortTestCase("UDPAddr with port above the 16-bit range",
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 65536}, netip.AddrPort{}, false),
	// A typed nil pointer is a recognised type carrying no address.
	newTypeSpecificAddrPortTestCase("nil AddrPort pointer", (*netip.AddrPort)(nil), netip.AddrPort{}, false),
	newTypeSpecificAddrPortTestCase("nil TCPAddr", (*net.TCPAddr)(nil), netip.AddrPort{}, false),
	newTypeSpecificAddrPortTestCase("nil UDPAddr", (*net.UDPAddr)(nil), netip.AddrPort{}, false),
	newTypeSpecificAddrPortTestCaseUnknown("unsupported type", "not an address"),
}

// newTypeSpecificAddrPortTestCase declares a row for a type the switch
// recognises, whether or not it yields an address.
func newTypeSpecificAddrPortTestCase(name string, input any, want netip.AddrPort,
	wantOK bool) typeSpecificAddrPortTestCase {
	return typeSpecificAddrPortTestCase{
		name:      name,
		input:     input,
		want:      want,
		wantOK:    wantOK,
		wantKnown: true,
	}
}

// newTypeSpecificAddrPortTestCaseUnknown declares a row for a type the switch
// does not recognise, which never yields an address.
func newTypeSpecificAddrPortTestCaseUnknown(name string, input any) typeSpecificAddrPortTestCase {
	return typeSpecificAddrPortTestCase{
		name:      name,
		input:     input,
		want:      netip.AddrPort{},
		wantOK:    false,
		wantKnown: false,
	}
}

func (tc typeSpecificAddrPortTestCase) Name() string {
	return tc.name
}

func (tc typeSpecificAddrPortTestCase) Test(t *testing.T) {
	t.Helper()

	got, ok, known := typeSpecificAddrPort(tc.input)
	AssertEqual(t, tc.wantKnown, known, "known")
	AssertEqual(t, tc.wantOK, ok, "ok")
	AssertEqual(t, tc.want, got, "address")
}

func TestTypeSpecificAddrPort(t *testing.T) {
	RunTestCases(t, typeSpecificAddrPortTestCases)
}
