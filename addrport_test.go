package core_test

import (
	"net"
	"net/netip"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that addrPortTestCase implements TestCase
var _ core.TestCase = addrPortTestCase{}

// Test cases for AddrPort
type addrPortTestCase struct {
	name   string
	input  any
	want   netip.AddrPort
	wantOK bool
}

var addrPortTestCases = []addrPortTestCase{
	// Direct netip.AddrPort types
	newAddrPortTestCase("direct AddrPort", netip.MustParseAddrPort("192.168.1.1:8080"),
		netip.MustParseAddrPort("192.168.1.1:8080"), true),
	newAddrPortTestCase("pointer to AddrPort", core.NewAddrPortPtrForTest(netip.MustParseAddrPort("10.0.0.1:443")),
		netip.MustParseAddrPort("10.0.0.1:443"), true),
	newAddrPortTestCase("zero AddrPort", netip.AddrPort{}, netip.AddrPort{}, false),
	newAddrPortTestCase("nil pointer to AddrPort", (*netip.AddrPort)(nil), netip.AddrPort{}, false),
	// TCP addresses
	newAddrPortTestCase("TCPAddr IPv4", &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080},
		netip.MustParseAddrPort("192.168.1.1:8080"), true),
	newAddrPortTestCase("TCPAddr IPv6", &net.TCPAddr{IP: net.ParseIP("2001:db8::1"), Port: 443},
		netip.MustParseAddrPort("[2001:db8::1]:443"), true),
	newAddrPortTestCase("TCPAddr with zone", &net.TCPAddr{IP: net.ParseIP("fe80::1"), Port: 22, Zone: "eth0"},
		netip.MustParseAddrPort("[fe80::1]:22"), true),
	newAddrPortTestCase("TCPAddr with nil IP", &net.TCPAddr{IP: nil, Port: 8080}, netip.AddrPort{}, false),
	newAddrPortTestCase("TCPAddr with port above the 16-bit range",
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 65536}, netip.AddrPort{}, false),
	newAddrPortTestCase("TCPAddr with zero port", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		netip.MustParseAddrPort("127.0.0.1:0"), true),
	newAddrPortTestCase("nil TCPAddr", (*net.TCPAddr)(nil), netip.AddrPort{}, false),
	// UDP addresses
	newAddrPortTestCase("UDPAddr IPv4", &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 53},
		netip.MustParseAddrPort("192.168.1.1:53"), true),
	newAddrPortTestCase("UDPAddr IPv6", &net.UDPAddr{IP: net.ParseIP("::1"), Port: 123},
		netip.MustParseAddrPort("[::1]:123"), true),
	newAddrPortTestCase("UDPAddr with nil IP", &net.UDPAddr{IP: nil, Port: 53}, netip.AddrPort{}, false),
	newAddrPortTestCase("nil UDPAddr", (*net.UDPAddr)(nil), netip.AddrPort{}, false),
	// Interface types
	newAddrPortTestCase("type with AddrPort() method", addrPortProvider{netip.MustParseAddrPort("127.0.0.1:9090")},
		netip.MustParseAddrPort("127.0.0.1:9090"), true),
	newAddrPortTestCase("type with AddrPort() method returning an invalid value",
		addrPortProvider{netip.AddrPort{}}, netip.AddrPort{}, false),
	newAddrPortTestCase("type with Addr() method returning TCPAddr",
		addrProvider{&net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 80}},
		netip.MustParseAddrPort("10.0.0.1:80"), true),
	newAddrPortTestCase("type with RemoteAddr() method",
		remoteAddrProvider{&net.UDPAddr{IP: net.ParseIP("8.8.8.8"), Port: 53}},
		netip.MustParseAddrPort("8.8.8.8:53"), true),
	newAddrPortTestCase("type with Addr() returning unsupported type",
		addrProvider{&net.UnixAddr{Name: "/tmp/socket", Net: "unix"}}, netip.AddrPort{}, false),
	// Invalid types
	newAddrPortTestCase("nil input", nil, netip.AddrPort{}, false),
	newAddrPortTestCase("string", "192.168.1.1:8080", netip.AddrPort{}, false),
	newAddrPortTestCase("int", 8080, netip.AddrPort{}, false),
	newAddrPortTestCase("net.IPAddr (no port)", &net.IPAddr{IP: net.ParseIP("192.168.1.1")}, netip.AddrPort{}, false),
	newAddrPortTestCase("empty struct", struct{}{}, netip.AddrPort{}, false),
}

// Types implementing various interfaces
type addrPortProvider struct {
	ap netip.AddrPort
}

func (p addrPortProvider) AddrPort() netip.AddrPort {
	return p.ap
}

type addrProvider struct {
	addr net.Addr
}

func (p addrProvider) Addr() net.Addr {
	return p.addr
}

type remoteAddrProvider struct {
	addr net.Addr
}

func (p remoteAddrProvider) RemoteAddr() net.Addr {
	return p.addr
}

func newAddrPortTestCase(name string, input any, want netip.AddrPort, wantOK bool) addrPortTestCase {
	return addrPortTestCase{
		name:   name,
		input:  input,
		want:   want,
		wantOK: wantOK,
	}
}

func (tc addrPortTestCase) Name() string {
	return tc.name
}

func (tc addrPortTestCase) Test(t *testing.T) {
	t.Helper()

	got, ok := core.AddrPort(tc.input)
	core.AssertEqual(t, tc.wantOK, ok, "ok")
	core.AssertEqual(t, tc.want, got, "address")
}

func TestAddrPort(t *testing.T) {
	core.RunTestCases(t, addrPortTestCases)
}

// Test edge cases and interface interactions
func TestAddrPortInterfaceChaining(t *testing.T) {
	t.Run("recursive interface resolution", testRecursiveInterfaceResolution)
	t.Run("nil from interface methods", testNilFromInterfaceMethods)
}

func testRecursiveInterfaceResolution(t *testing.T) {
	t.Helper()

	// Create a provider that returns another provider
	tcpAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080}
	provider := addrProvider{tcpAddr}

	got, ok := core.AddrPort(provider)
	core.AssertTrue(t, ok, "ok")

	want := netip.MustParseAddrPort("192.168.1.1:8080")
	core.AssertEqual(t, want, got, "address")
}

func testNilFromInterfaceMethods(t *testing.T) {
	t.Helper()

	nilProvider := addrProvider{nil}
	got, ok := core.AddrPort(nilProvider)
	core.AssertFalse(t, ok, "nil Addr() ok")
	core.AssertEqual(t, netip.AddrPort{}, got, "nil Addr() address")

	nilRemoteProvider := remoteAddrProvider{nil}
	got, ok = core.AddrPort(nilRemoteProvider)
	core.AssertFalse(t, ok, "nil RemoteAddr() ok")
	core.AssertEqual(t, netip.AddrPort{}, got, "nil RemoteAddr() address")
}

// Test with real network connection types (mock)
type mockConn struct {
	remote net.Addr
}

func (c mockConn) RemoteAddr() net.Addr {
	return c.remote
}

func TestAddrPortWithMockConnection(t *testing.T) {
	conn := mockConn{
		remote: &net.TCPAddr{IP: net.ParseIP("192.168.1.100"), Port: 12345},
	}

	got, ok := core.AddrPort(conn)
	core.AssertTrue(t, ok, "ok")

	want := netip.MustParseAddrPort("192.168.1.100:12345")
	core.AssertEqual(t, want, got, "address")
}

// Test with IPv4 addresses using To4()
func TestAddrPortIPv4Handling(t *testing.T) {
	// Test with explicit IPv4
	ipv4 := net.ParseIP("192.168.1.1").To4()
	tcpAddr := &net.TCPAddr{IP: ipv4, Port: 8080}

	got, ok := core.AddrPort(tcpAddr)
	core.AssertTrue(t, ok, "ok")

	// Should get the IPv4 address directly
	want := netip.MustParseAddrPort("192.168.1.1:8080")
	core.AssertEqual(t, want, got, "address")
}

// Benchmarks
func BenchmarkAddrPortDirect(b *testing.B) {
	ap := netip.MustParseAddrPort("192.168.1.1:8080")
	for b.Loop() {
		_, _ = core.AddrPort(ap)
	}
}

func BenchmarkAddrPortPointer(b *testing.B) {
	ap := core.NewAddrPortPtrForTest(netip.MustParseAddrPort("192.168.1.1:8080"))
	for b.Loop() {
		_, _ = core.AddrPort(ap)
	}
}

func BenchmarkAddrPortTCPAddr(b *testing.B) {
	addr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080}
	for b.Loop() {
		_, _ = core.AddrPort(addr)
	}
}

func BenchmarkAddrPortInterface(b *testing.B) {
	provider := addrPortProvider{netip.MustParseAddrPort("192.168.1.1:8080")}
	for b.Loop() {
		_, _ = core.AddrPort(provider)
	}
}
