package core

import (
	"net"
	"net/netip"
	"reflect"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ TestCase = parseAddrTestCase{}
var _ TestCase = parseNetIPTestCase{}
var _ TestCase = addrFromNetIPTestCase{}
var _ TestCase = getInterfacesNamesTestCase{}

// Test cases for ParseAddr
type parseAddrTestCase struct {
	// netip.Addr struct - result value
	want netip.Addr

	// String fields - alphabetically ordered
	input string
	name  string

	// Boolean field (1 byte) - flags
	wantErr bool
}

var parseAddrTestCases = []parseAddrTestCase{
	newParseAddrTestCaseStr("IPv4 address", "192.168.1.1", "192.168.1.1", false),
	newParseAddrTestCaseStr("IPv6 address", "2001:db8::1", "2001:db8::1", false),
	newParseAddrTestCase("IPv4 unspecified shorthand", "0", netip.IPv4Unspecified(), false),
	newParseAddrTestCase("IPv6 unspecified", "::", netip.IPv6Unspecified(), false),
	newParseAddrTestCaseStr("IPv4 loopback", "127.0.0.1", "127.0.0.1", false),
	newParseAddrTestCaseStr("IPv6 loopback", "::1", "::1", false),
	newParseAddrTestCaseStr("Invalid address", "not-an-ip", "", true),
	newParseAddrTestCaseStr("Empty string", "", "", true),
	newParseAddrTestCaseStr("IPv4 with port", "192.168.1.1:8080", "", true),
	newParseAddrTestCaseStr("IPv6 with zone", "fe80::1%eth0", "fe80::1%eth0", false),
}

func newParseAddrTestCase(name, input string, want netip.Addr, wantErr bool) parseAddrTestCase {
	return parseAddrTestCase{
		name:    name,
		input:   input,
		want:    want,
		wantErr: wantErr,
	}
}

//revive:disable-next-line:flag-parameter
func newParseAddrTestCaseStr(name, input, wantAddr string, wantErr bool) parseAddrTestCase {
	var want netip.Addr
	if !wantErr && wantAddr != "" {
		want = netip.MustParseAddr(wantAddr)
	}
	return parseAddrTestCase{
		name:    name,
		input:   input,
		want:    want,
		wantErr: wantErr,
	}
}

func (tc parseAddrTestCase) Name() string {
	return tc.name
}

func (tc parseAddrTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := ParseAddr(tc.input)
	if tc.wantErr {
		if err == nil {
			t.Error("Expected error but got nil")
		}
	} else {
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
		if got != tc.want {
			t.Errorf("Expected %v, got %v", tc.want, got)
		}
	}
}

func TestParseAddr(t *testing.T) {
	RunTestCases(t, parseAddrTestCases)
}

// Test cases for ParseNetIP
type parseNetIPTestCase struct {
	name    string
	input   string
	want    net.IP
	wantErr bool
}

var parseNetIPTestCases = []parseNetIPTestCase{
	newParseNetIPTestCase("IPv4 address", "192.168.1.1", net.ParseIP("192.168.1.1"), false),
	newParseNetIPTestCase("IPv6 address", "2001:db8::1", net.ParseIP("2001:db8::1"), false),
	newParseNetIPTestCase("IPv4 unspecified shorthand", "0", net.IPv4zero, false),
	newParseNetIPTestCase("IPv6 unspecified", "::", net.IPv6zero, false),
	newParseNetIPTestCase("Invalid address", "invalid", nil, true),
	newParseNetIPTestCase("Empty string", "", nil, true),
}

func newParseNetIPTestCase(name, input string, want net.IP, wantErr bool) parseNetIPTestCase {
	return parseNetIPTestCase{
		name:    name,
		input:   input,
		want:    want,
		wantErr: wantErr,
	}
}

func (tc parseNetIPTestCase) Name() string {
	return tc.name
}

func (tc parseNetIPTestCase) Test(t *testing.T) {
	t.Helper()

	got, err := ParseNetIP(tc.input)
	if tc.wantErr {
		if err == nil {
			t.Error("Expected error but got nil")
		}
	} else {
		if err != nil {
			t.Errorf("Expected no error but got: %v", err)
		}
		if !got.Equal(tc.want) {
			t.Errorf("Expected %v, got %v", tc.want, got)
		}
	}
}

func TestParseNetIP(t *testing.T) {
	RunTestCases(t, parseNetIPTestCases)
}

// Test cases for AddrFromNetIP
type addrFromNetIPTestCase struct {
	// netip.Addr struct - result value
	want netip.Addr

	// Interface field - input value
	input net.Addr

	// String field - test name
	name string

	// Boolean field (1 byte) - success flag
	ok bool
}

var addrFromNetIPTestCases = []addrFromNetIPTestCase{
	newAddrFromNetIPTestCase("IPAddr IPv4", &net.IPAddr{IP: net.ParseIP("192.168.1.1")},
		netip.MustParseAddr("192.168.1.1"), true),
	newAddrFromNetIPTestCase("IPAddr IPv6", &net.IPAddr{IP: net.ParseIP("2001:db8::1")},
		netip.MustParseAddr("2001:db8::1"), true),
	newAddrFromNetIPTestCase("IPNet IPv4", &net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)},
		netip.MustParseAddr("10.0.0.0"), true),
	newAddrFromNetIPTestCase("IPNet IPv6", &net.IPNet{IP: net.ParseIP("2001:db8::"), Mask: net.CIDRMask(64, 128)},
		netip.MustParseAddr("2001:db8::"), true),
	newAddrFromNetIPTestCase("TCPAddr", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, netip.Addr{}, false),
	newAddrFromNetIPTestCase("UDPAddr", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, netip.Addr{}, false),
	newAddrFromNetIPTestCase("nil input", nil, netip.Addr{}, false),
	newAddrFromNetIPTestCase("IPAddr with nil IP", &net.IPAddr{IP: nil}, netip.Addr{}, false),
	newAddrFromNetIPTestCase("IPNet with nil IP", &net.IPNet{IP: nil, Mask: net.CIDRMask(24, 32)}, netip.Addr{}, false),
}

func newAddrFromNetIPTestCase(name string, input net.Addr, want netip.Addr, ok bool) addrFromNetIPTestCase {
	return addrFromNetIPTestCase{
		name:  name,
		input: input,
		want:  want,
		ok:    ok,
	}
}

func (tc addrFromNetIPTestCase) Name() string {
	return tc.name
}

func (tc addrFromNetIPTestCase) Test(t *testing.T) {
	t.Helper()

	got, ok := AddrFromNetIP(tc.input)
	if ok != tc.ok {
		t.Errorf("Expected ok=%v, got %v", tc.ok, ok)
	}
	if ok && got != tc.want {
		t.Errorf("Expected %v, got %v", tc.want, got)
	}
}

func TestAddrFromNetIP(t *testing.T) {
	RunTestCases(t, addrFromNetIPTestCases)
}

// Test GetStringIPAddresses
func TestGetStringIPAddresses(t *testing.T) {
	// Since we can't easily mock net.InterfaceByName and net.InterfaceAddrs,
	// we'll test with the real interfaces but handle the case where there are none

	t.Run("all interfaces", testAllStringInterfaces)

	t.Run("specific interface", testSpecificInvalidInterface)

	t.Run("loopback interface", testLoopbackInterface)
}

func testSpecificInvalidInterface(t *testing.T) {
	t.Helper()
	// Try with a non-existent interface
	_, err := GetStringIPAddresses("invalid-interface-name")
	if err == nil {
		t.Error("Expected error for invalid interface")
	}
}

func testLoopbackInterface(t *testing.T) {
	// Most systems have a loopback interface
	loopbackNames := S("lo", "lo0", "Loopback Pseudo-Interface 1")

	var found bool
	for _, name := range loopbackNames {
		addrs, err := GetStringIPAddresses(name)
		if err == nil {
			found = true
			// Should have at least one address (127.0.0.1 or ::1)
			if len(addrs) == 0 {
				t.Errorf("Expected at least one address for loopback interface %s", name)
			}
			break
		}
	}

	if !found {
		t.Skip("No loopback interface found on this system")
	}
}

func testAllStringInterfaces(t *testing.T) {
	t.Helper()
	addrs, err := GetStringIPAddresses()
	if err != nil {
		// Some systems might not have any interfaces
		t.Logf("Got error (possibly no interfaces): %v", err)
		return
	}

	// At least check it returns a slice (could be empty)
	if addrs == nil {
		t.Error("Expected non-nil slice")
	}

	// Verify all returned addresses are valid strings
	for _, addr := range addrs {
		if _, err := netip.ParseAddr(addr); err != nil {
			t.Errorf("Invalid address string: %s", addr)
		}
	}
}

// Test GetNetIPAddresses
func TestGetNetIPAddresses(t *testing.T) {
	t.Run("all interfaces", testAllNetIPInterfaces)
	t.Run("invalid interface", testInvalidNetIPInterface)
}

func testAllNetIPInterfaces(t *testing.T) {
	t.Helper()
	addrs, err := GetNetIPAddresses()
	if err != nil {
		// Some systems might not have any interfaces
		t.Logf("Got error (possibly no interfaces): %v", err)
		return
	}

	// At least check it returns a slice (could be empty)
	if addrs == nil {
		t.Error("Expected non-nil slice")
	}

	// Verify all returned addresses are valid net.IP
	for _, addr := range addrs {
		if len(addr) == 0 {
			t.Error("Got nil or empty net.IP")
		}
	}
}

func testInvalidNetIPInterface(t *testing.T) {
	t.Helper()
	_, err := GetNetIPAddresses("invalid-interface-name")
	if err == nil {
		t.Error("Expected error for invalid interface")
	}
}

// Test GetIPAddresses
func TestGetIPAddresses(t *testing.T) {
	t.Run("all interfaces", testAllIPAddressInterfaces)
	t.Run("multiple interfaces with error", testMultipleInterfacesWithError)
}

func testAllIPAddressInterfaces(t *testing.T) {
	t.Helper()
	addrs, err := GetIPAddresses()
	if err != nil {
		// Some systems might not have any interfaces
		t.Logf("Got error (possibly no interfaces): %v", err)
		return
	}

	// At least check it returns a slice (could be empty)
	if addrs == nil {
		t.Error("Expected non-nil slice")
	}

	// Verify all returned addresses are valid
	for _, addr := range addrs {
		if !addr.IsValid() {
			t.Error("Got invalid netip.Addr")
		}
	}
}

func testMultipleInterfacesWithError(t *testing.T) {
	t.Helper()
	// Try with one valid and one invalid interface
	ifaces, err := GetInterfacesNames()
	if err != nil || len(ifaces) == 0 {
		t.Skip("No interfaces available for testing")
	}

	// Mix valid and invalid interface names
	_, err = GetIPAddresses(ifaces[0], "invalid-interface-name")
	if err == nil {
		t.Error("Expected error when one interface doesn't exist")
	}
}

// Test GetInterfacesNames
type getInterfacesNamesTestCase struct {
	name   string
	except []string
}

var getInterfacesNamesTestCases = []getInterfacesNamesTestCase{
	newGetInterfacesNamesTestCase("no exclusions", nil),
	newGetInterfacesNamesTestCase("empty exclusions", S[string]()),
	newGetInterfacesNamesTestCase("exclude invalid", S("invalid-interface")),
	newGetInterfacesNamesTestCase("exclude multiple", S("eth0", "eth1", "lo")),
}

func newGetInterfacesNamesTestCase(name string, except []string) getInterfacesNamesTestCase {
	return getInterfacesNamesTestCase{
		name:   name,
		except: except,
	}
}

func (tc getInterfacesNamesTestCase) Name() string {
	return tc.name
}

func (tc getInterfacesNamesTestCase) Test(t *testing.T) {
	t.Helper()

	names, err := GetInterfacesNames(tc.except...)
	AssertNoError(t, err, "GetInterfacesNames")
	AssertNotNil(t, names, "interface names slice")

	// Check that excluded names are not in the result
	for _, excluded := range tc.except {
		AssertFalse(t, SliceContains(names, excluded), "contains interface %q", excluded)
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, name := range names {
		AssertFalse(t, seen[name], "duplicate interface %q", name)
		seen[name] = true
	}
}

func TestGetInterfacesNames(t *testing.T) {
	RunTestCases(t, getInterfacesNamesTestCases)

	// Additional test: verify exclusion actually works
	t.Run("exclusion removes interfaces", testExclusionRemovesInterfaces)
}

func testExclusionRemovesInterfaces(t *testing.T) {
	t.Helper()
	all, err := GetInterfacesNames()
	if err != nil || len(all) == 0 {
		t.Skip("No interfaces available for testing")
	}

	// Exclude the first interface
	excluded := all[0]
	filtered, err := GetInterfacesNames(excluded)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Should have one less interface
	if len(filtered) != len(all)-1 {
		t.Errorf("Expected %d interfaces, got %d", len(all)-1, len(filtered))
	}

	// Verify the excluded interface is not present
	for _, name := range filtered {
		if name == excluded {
			t.Errorf("Found excluded interface %s in result", excluded)
		}
	}
}

// Test internal helper functions
func TestAsStringIPAddresses(t *testing.T) {
	addrs := S[netip.Addr](
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("2001:db8::1"),
		netip.Addr{}, // Invalid address
		netip.MustParseAddr("10.0.0.1"),
	)

	result := asStringIPAddresses(addrs...)
	expected := S("192.168.1.1", "2001:db8::1", "10.0.0.1")

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestAsNetIPAddresses(t *testing.T) {
	addrs := S[netip.Addr](
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("2001:db8::1"),
		netip.Addr{},                              // Invalid address
		netip.MustParseAddr("::ffff:192.168.1.2"), // IPv4-mapped IPv6
	)

	result := asNetIPAddresses(addrs...)

	// Should have 3 valid addresses (invalid one skipped)
	if len(result) != 3 {
		t.Errorf("Expected 3 addresses, got %d", len(result))
	}

	// Verify the IPv4-mapped address is unmapped
	lastAddr := result[2]
	if !lastAddr.Equal(net.ParseIP("192.168.1.2")) {
		t.Errorf("Expected unmapped IPv4 address, got %v", lastAddr)
	}
}

func TestAsNetIP(t *testing.T) {
	t.Run("IPv4", testAsNetIPv4)
	t.Run("IPv6", testAsNetIPv6)
}

func testAsNetIPv4(t *testing.T) {
	t.Helper()
	addr := netip.MustParseAddr("192.168.1.1")
	ip := asNetIP(addr)
	expected := net.ParseIP("192.168.1.1").To4()
	if !ip.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, ip)
	}
	// Should be 4 bytes for IPv4
	if len(ip) != 4 {
		t.Errorf("Expected 4 bytes for IPv4, got %d", len(ip))
	}
}

func testAsNetIPv6(t *testing.T) {
	t.Helper()
	addr := netip.MustParseAddr("2001:db8::1")
	ip := asNetIP(addr)
	expected := net.ParseIP("2001:db8::1")
	if !ip.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, ip)
	}
	// Should be 16 bytes for IPv6
	if len(ip) != 16 {
		t.Errorf("Expected 16 bytes for IPv6, got %d", len(ip))
	}
}

func TestAppendNetIPAsIP(t *testing.T) {
	var out []netip.Addr

	addrs := S[net.Addr](
		&net.IPAddr{IP: net.ParseIP("192.168.1.1")},
		&net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)},
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, // Should be skipped
		&net.IPAddr{IP: nil}, // Should be skipped
	)

	result := appendNetIPAsIP(out, addrs...)

	// Should have 2 valid addresses
	if len(result) != 2 {
		t.Errorf("Expected 2 addresses, got %d", len(result))
	}

	// Verify the addresses
	if result[0] != netip.MustParseAddr("192.168.1.1") {
		t.Errorf("Expected first address to be 192.168.1.1, got %v", result[0])
	}
	if result[1] != netip.MustParseAddr("10.0.0.0") {
		t.Errorf("Expected second address to be 10.0.0.0, got %v", result[1])
	}
}

// Benchmarks
func BenchmarkParseAddr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseAddr("192.168.1.1")
	}
}

func BenchmarkParseNetIP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ParseNetIP("192.168.1.1")
	}
}

func BenchmarkGetStringIPAddresses(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetStringIPAddresses()
	}
}

func BenchmarkAddrFromNetIP(b *testing.B) {
	addr := &net.IPAddr{IP: net.ParseIP("192.168.1.1")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AddrFromNetIP(addr)
	}
}
