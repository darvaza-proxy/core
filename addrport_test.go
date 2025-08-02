package core

import (
	"net"
	"net/netip"
	"testing"
)

// Test cases for AddrPort
type addrPortTestCase struct {
	name   string
	input  any
	want   netip.AddrPort
	wantOK bool
}

var addrPortTestCases = []addrPortTestCase{
	// Direct netip.AddrPort types
	{
		name:   "direct AddrPort",
		input:  netip.MustParseAddrPort("192.168.1.1:8080"),
		want:   netip.MustParseAddrPort("192.168.1.1:8080"),
		wantOK: true,
	},
	{
		name:   "pointer to AddrPort",
		input:  addrPortPtr(netip.MustParseAddrPort("10.0.0.1:443")),
		want:   netip.MustParseAddrPort("10.0.0.1:443"),
		wantOK: true,
	},
	{
		name:   "zero AddrPort",
		input:  netip.AddrPort{},
		want:   netip.AddrPort{},
		wantOK: true,
	},
	// TCP addresses
	{
		name:   "TCPAddr IPv4",
		input:  &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 8080},
		want:   netip.MustParseAddrPort("192.168.1.1:8080"),
		wantOK: true,
	},
	{
		name:   "TCPAddr IPv6",
		input:  &net.TCPAddr{IP: net.ParseIP("2001:db8::1"), Port: 443},
		want:   netip.MustParseAddrPort("[2001:db8::1]:443"),
		wantOK: true,
	},
	{
		name:   "TCPAddr with zone",
		input:  &net.TCPAddr{IP: net.ParseIP("fe80::1"), Port: 22, Zone: "eth0"},
		want:   netip.MustParseAddrPort("[fe80::1]:22"),
		wantOK: true,
	},
	{
		name:   "TCPAddr with nil IP",
		input:  &net.TCPAddr{IP: nil, Port: 8080},
		want:   netip.AddrPort{},
		wantOK: false,
	},
	{
		name:   "TCPAddr with zero port",
		input:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0},
		want:   netip.MustParseAddrPort("127.0.0.1:0"),
		wantOK: true,
	},
	// UDP addresses
	{
		name:   "UDPAddr IPv4",
		input:  &net.UDPAddr{IP: net.ParseIP("192.168.1.1"), Port: 53},
		want:   netip.MustParseAddrPort("192.168.1.1:53"),
		wantOK: true,
	},
	{
		name:   "UDPAddr IPv6",
		input:  &net.UDPAddr{IP: net.ParseIP("::1"), Port: 123},
		want:   netip.MustParseAddrPort("[::1]:123"),
		wantOK: true,
	},
	{
		name:   "UDPAddr with nil IP",
		input:  &net.UDPAddr{IP: nil, Port: 53},
		want:   netip.AddrPort{},
		wantOK: false,
	},
	// Interface types
	{
		name:   "type with AddrPort() method",
		input:  addrPortProvider{netip.MustParseAddrPort("127.0.0.1:9090")},
		want:   netip.MustParseAddrPort("127.0.0.1:9090"),
		wantOK: true,
	},
	{
		name:   "type with Addr() method returning TCPAddr",
		input:  addrProvider{&net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 80}},
		want:   netip.MustParseAddrPort("10.0.0.1:80"),
		wantOK: true,
	},
	{
		name:   "type with RemoteAddr() method",
		input:  remoteAddrProvider{&net.UDPAddr{IP: net.ParseIP("8.8.8.8"), Port: 53}},
		want:   netip.MustParseAddrPort("8.8.8.8:53"),
		wantOK: true,
	},
	{
		name:   "type with Addr() returning unsupported type",
		input:  addrProvider{&net.UnixAddr{Name: "/tmp/socket", Net: "unix"}},
		want:   netip.AddrPort{},
		wantOK: false,
	},
	// Invalid types
	{
		name:   "nil input",
		input:  nil,
		want:   netip.AddrPort{},
		wantOK: false,
	},
	{
		name:   "string",
		input:  "192.168.1.1:8080",
		want:   netip.AddrPort{},
		wantOK: false,
	},
	{
		name:   "int",
		input:  8080,
		want:   netip.AddrPort{},
		wantOK: false,
	},
	{
		name:   "net.IPAddr (no port)",
		input:  &net.IPAddr{IP: net.ParseIP("192.168.1.1")},
		want:   netip.AddrPort{},
		wantOK: false,
	},
	{
		name:   "empty struct",
		input:  struct{}{},
		want:   netip.AddrPort{},
		wantOK: false,
	},
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

func (tc addrPortTestCase) test(t *testing.T) {
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
	for _, tc := range addrPortTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test typeSpecificAddrPort directly
type typeSpecificAddrPortTestCase struct {
	name   string
	input  any
	want   netip.AddrPort
	wantOK bool
}

var typeSpecificAddrPortTestCases = []typeSpecificAddrPortTestCase{
	{
		name:   "AddrPort value",
		input:  netip.MustParseAddrPort("192.168.1.1:8080"),
		want:   netip.MustParseAddrPort("192.168.1.1:8080"),
		wantOK: true,
	},
	{
		name:   "AddrPort pointer",
		input:  addrPortPtr(netip.MustParseAddrPort("10.0.0.1:443")),
		want:   netip.MustParseAddrPort("10.0.0.1:443"),
		wantOK: true,
	},
	{
		name:   "TCPAddr",
		input:  &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9000},
		want:   netip.MustParseAddrPort("127.0.0.1:9000"),
		wantOK: true,
	},
	{
		name:   "UDPAddr",
		input:  &net.UDPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 53},
		want:   netip.MustParseAddrPort("10.0.0.1:53"),
		wantOK: true,
	},
	{
		name:   "TCPAddr with invalid IP",
		input:  &net.TCPAddr{IP: S[byte](1, 2, 3), Port: 80}, // Invalid IP length
		want:   netip.AddrPort{},
		wantOK: false,
	},
	{
		name:   "UDPAddr with invalid IP",
		input:  &net.UDPAddr{IP: S[byte](), Port: 80}, // Empty IP
		want:   netip.AddrPort{},
		wantOK: false,
	},
	{
		name:   "unsupported type",
		input:  "not an address",
		want:   netip.AddrPort{},
		wantOK: false,
	},
}

func (tc typeSpecificAddrPortTestCase) test(t *testing.T) {
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
	for _, tc := range typeSpecificAddrPortTestCases {
		t.Run(tc.name, tc.test)
	}
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
