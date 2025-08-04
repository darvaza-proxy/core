package core

import (
	"net"
	"net/netip"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ TestCase = addrPortTestCase{}
var _ TestCase = typeSpecificAddrPortTestCase{}

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
	newAddrPortTestCase("pointer to AddrPort", addrPortPtr(netip.MustParseAddrPort("10.0.0.1:443")),
		netip.MustParseAddrPort("10.0.0.1:443"), true),
	newAddrPortTestCase("zero AddrPort", netip.AddrPort{}, netip.AddrPort{}, true),
	// TCP addresses
	newAddrPortTestCase("TCPAddr IPv4", &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080},
		netip.MustParseAddrPort("192.168.1.1:8080"), true),
	newAddrPortTestCase("TCPAddr IPv6", &net.TCPAddr{IP: net.ParseIP("2001:db8::1"), Port: 443},
		netip.MustParseAddrPort("[2001:db8::1]:443"), true),
	newAddrPortTestCase("TCPAddr with zone", &net.TCPAddr{IP: net.ParseIP("fe80::1"), Port: 22, Zone: "eth0"},
		netip.MustParseAddrPort("[fe80::1]:22"), true),
	newAddrPortTestCase("TCPAddr with nil IP", &net.TCPAddr{IP: nil, Port: 8080}, netip.AddrPort{}, false),
	newAddrPortTestCase("TCPAddr with zero port", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		netip.MustParseAddrPort("127.0.0.1:0"), true),
	// UDP addresses
	newAddrPortTestCase("UDPAddr IPv4", &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 53},
		netip.MustParseAddrPort("192.168.1.1:53"), true),
	newAddrPortTestCase("UDPAddr IPv6", &net.UDPAddr{IP: net.ParseIP("::1"), Port: 123},
		netip.MustParseAddrPort("[::1]:123"), true),
	newAddrPortTestCase("UDPAddr with nil IP", &net.UDPAddr{IP: nil, Port: 53}, netip.AddrPort{}, false),
	// Interface types
	newAddrPortTestCase("type with AddrPort() method", addrPortProvider{netip.MustParseAddrPort("127.0.0.1:9090")},
		netip.MustParseAddrPort("127.0.0.1:9090"), true),
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

// Helper to create pointer to AddrPort
func addrPortPtr(ap netip.AddrPort) *netip.AddrPort {
	return &ap
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

	got, ok := AddrPort(tc.input)
	if ok != tc.wantOK {
		t.Errorf("Expected ok=%v, got %v", tc.wantOK, ok)
	}
	if ok && got != tc.want {
		t.Errorf("Expected %v, got %v", tc.want, got)
	}
}

func TestAddrPort(t *testing.T) {
	RunTestCases(t, addrPortTestCases)
}

// Test typeSpecificAddrPort directly
type typeSpecificAddrPortTestCase struct {
	name   string
	input  any
	want   netip.AddrPort
	wantOK bool
}

var typeSpecificAddrPortTestCases = []typeSpecificAddrPortTestCase{
	newTypeSpecificAddrPortTestCase("AddrPort value", netip.MustParseAddrPort("192.168.1.1:8080"),
		netip.MustParseAddrPort("192.168.1.1:8080"), true),
	newTypeSpecificAddrPortTestCase("AddrPort pointer", addrPortPtr(netip.MustParseAddrPort("10.0.0.1:443")),
		netip.MustParseAddrPort("10.0.0.1:443"), true),
	newTypeSpecificAddrPortTestCase("TCPAddr", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9000},
		netip.MustParseAddrPort("127.0.0.1:9000"), true),
	newTypeSpecificAddrPortTestCase("UDPAddr", &net.UDPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 53},
		netip.MustParseAddrPort("10.0.0.1:53"), true),
	newTypeSpecificAddrPortTestCase("TCPAddr with invalid IP", &net.TCPAddr{IP: S[byte](1, 2, 3), Port: 80},
		netip.AddrPort{}, false), // Invalid IP length
	newTypeSpecificAddrPortTestCase("UDPAddr with invalid IP", &net.UDPAddr{IP: S[byte](), Port: 80},
		netip.AddrPort{}, false), // Empty IP
	newTypeSpecificAddrPortTestCase("unsupported type", "not an address", netip.AddrPort{}, false),
}

func newTypeSpecificAddrPortTestCase(name string, input any, want netip.AddrPort,
	wantOK bool) typeSpecificAddrPortTestCase {
	return typeSpecificAddrPortTestCase{
		name:   name,
		input:  input,
		want:   want,
		wantOK: wantOK,
	}
}

func (tc typeSpecificAddrPortTestCase) Name() string {
	return tc.name
}

func (tc typeSpecificAddrPortTestCase) Test(t *testing.T) {
	t.Helper()

	got, ok := typeSpecificAddrPort(tc.input)
	if ok != tc.wantOK {
		t.Errorf("Expected ok=%v, got %v", tc.wantOK, ok)
	}
	if ok && got != tc.want {
		t.Errorf("Expected %v, got %v", tc.want, got)
	}
}

func TestTypeSpecificAddrPort(t *testing.T) {
	RunTestCases(t, typeSpecificAddrPortTestCases)
}

// Test edge cases and interface interactions
func TestAddrPortInterfaceChaining(t *testing.T) {
	t.Run("recursive interface resolution", testRecursiveInterfaceResolution)
	t.Run("nil from interface methods", testNilFromInterfaceMethods)
}

func testRecursiveInterfaceResolution(t *testing.T) {
	// Create a provider that returns another provider
	tcpAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080}
	provider := addrProvider{tcpAddr}

	got, ok := AddrPort(provider)
	if !ok {
		t.Error("Expected success for addrProvider with TCPAddr")
	}

	want := netip.MustParseAddrPort("192.168.1.1:8080")
	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

func testNilFromInterfaceMethods(t *testing.T) {
	nilProvider := addrProvider{nil}
	_, ok := AddrPort(nilProvider)
	if ok {
		t.Error("Expected failure for nil Addr()")
	}

	nilRemoteProvider := remoteAddrProvider{nil}
	_, ok = AddrPort(nilRemoteProvider)
	if ok {
		t.Error("Expected failure for nil RemoteAddr()")
	}
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

	got, ok := AddrPort(conn)
	if !ok {
		t.Error("Expected success for mock connection")
	}

	want := netip.MustParseAddrPort("192.168.1.100:12345")
	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

// Test with IPv4 addresses using To4()
func TestAddrPortIPv4Handling(t *testing.T) {
	// Test with explicit IPv4
	ipv4 := net.ParseIP("192.168.1.1").To4()
	tcpAddr := &net.TCPAddr{IP: ipv4, Port: 8080}

	got, ok := AddrPort(tcpAddr)
	if !ok {
		t.Error("Expected success for IPv4 address")
	}

	// Should get the IPv4 address directly
	want := netip.MustParseAddrPort("192.168.1.1:8080")
	if got != want {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

// Benchmarks
func BenchmarkAddrPortDirect(b *testing.B) {
	ap := netip.MustParseAddrPort("192.168.1.1:8080")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddrPort(ap)
	}
}

func BenchmarkAddrPortPointer(b *testing.B) {
	ap := addrPortPtr(netip.MustParseAddrPort("192.168.1.1:8080"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddrPort(ap)
	}
}

func BenchmarkAddrPortTCPAddr(b *testing.B) {
	addr := &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddrPort(addr)
	}
}

func BenchmarkAddrPortInterface(b *testing.B) {
	provider := addrPortProvider{netip.MustParseAddrPort("192.168.1.1:8080")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddrPort(provider)
	}
}
