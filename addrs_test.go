package core_test

import (
	"net"
	"net/netip"
	"testing"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement TestCase interface
var (
	_ core.TestCase = parseAddrTestCase{}
	_ core.TestCase = parseNetIPTestCase{}
	_ core.TestCase = addrFromNetIPTestCase{}
	_ core.TestCase = getInterfacesNamesTestCase{}
)

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

	got, err := core.ParseAddr(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "error")
		return
	}

	core.AssertNoError(t, err, "parse")
	core.AssertEqual(t, tc.want, got, "address")
}

func TestParseAddr(t *testing.T) {
	core.RunTestCases(t, parseAddrTestCases)
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

	got, err := core.ParseNetIP(tc.input)
	if tc.wantErr {
		core.AssertError(t, err, "error")
		return
	}

	core.AssertNoError(t, err, "parse")
	core.AssertEqual(t, tc.want, got, "IP")
}

func TestParseNetIP(t *testing.T) {
	core.RunTestCases(t, parseNetIPTestCases)
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

	got, ok := core.AddrFromNetIP(tc.input)
	core.AssertEqual(t, tc.ok, ok, "ok")
	core.AssertEqual(t, tc.want, got, "address")
}

func TestAddrFromNetIP(t *testing.T) {
	core.RunTestCases(t, addrFromNetIPTestCases)
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
	_, err := core.GetStringIPAddresses("invalid-interface-name")
	core.AssertError(t, err, "invalid interface")
}

func testLoopbackInterface(t *testing.T) {
	t.Helper()
	// Most systems have a loopback interface
	loopbackNames := core.S("lo", "lo0", "Loopback Pseudo-Interface 1")

	var found bool
	for _, name := range loopbackNames {
		addrs, err := core.GetStringIPAddresses(name)
		if err == nil {
			found = true
			// Should have at least one address (127.0.0.1 or ::1)
			core.AssertTrue(t, len(addrs) > 0, "loopback addresses")
			break
		}
	}

	if !found {
		t.Skip("No loopback interface found on this system")
	}
}

func testAllStringInterfaces(t *testing.T) {
	t.Helper()
	addrs, err := core.GetStringIPAddresses()
	if err != nil {
		// Some systems might not have any interfaces
		t.Logf("Got error (possibly no interfaces): %v", err)
		return
	}

	// At least check it returns a slice (could be empty)
	core.AssertNotNil(t, addrs, "addresses")

	// Verify all returned addresses are valid strings
	for _, addr := range addrs {
		_, err := netip.ParseAddr(addr)
		core.AssertNoError(t, err, "parse")
	}
}

// Test GetNetIPAddresses
func TestGetNetIPAddresses(t *testing.T) {
	t.Run("all interfaces", testAllNetIPInterfaces)
	t.Run("invalid interface", testInvalidNetIPInterface)
}

func testAllNetIPInterfaces(t *testing.T) {
	t.Helper()
	addrs, err := core.GetNetIPAddresses()
	if err != nil {
		// Some systems might not have any interfaces
		t.Logf("Got error (possibly no interfaces): %v", err)
		return
	}

	// At least check it returns a slice (could be empty)
	core.AssertNotNil(t, addrs, "addresses")

	// Verify all returned addresses are valid net.IP
	for _, addr := range addrs {
		core.AssertTrue(t, len(addr) > 0, "net.IP length")
	}
}

func testInvalidNetIPInterface(t *testing.T) {
	t.Helper()
	_, err := core.GetNetIPAddresses("invalid-interface-name")
	core.AssertError(t, err, "invalid interface")
}

// Test GetIPAddresses
func TestGetIPAddresses(t *testing.T) {
	t.Run("all interfaces", testAllIPAddressInterfaces)
	t.Run("multiple interfaces with error", testMultipleInterfacesWithError)
}

func testAllIPAddressInterfaces(t *testing.T) {
	t.Helper()
	addrs, err := core.GetIPAddresses()
	if err != nil {
		// Some systems might not have any interfaces
		t.Logf("Got error (possibly no interfaces): %v", err)
		return
	}

	// At least check it returns a slice (could be empty)
	core.AssertNotNil(t, addrs, "addresses")

	// Verify all returned addresses are valid
	for _, addr := range addrs {
		core.AssertTrue(t, addr.IsValid(), "address validity")
	}
}

func testMultipleInterfacesWithError(t *testing.T) {
	t.Helper()
	// Try with one valid and one invalid interface
	ifaces, err := core.GetInterfacesNames()
	if err != nil || len(ifaces) == 0 {
		t.Skip("No interfaces available for testing")
	}

	// Mix valid and invalid interface names
	_, err = core.GetIPAddresses(ifaces[0], "invalid-interface-name")
	core.AssertError(t, err, "mixed valid and invalid interfaces")
}

// Test GetInterfacesNames
type getInterfacesNamesTestCase struct {
	name   string
	except []string
}

var getInterfacesNamesTestCases = []getInterfacesNamesTestCase{
	newGetInterfacesNamesTestCase("no exclusions", nil),
	newGetInterfacesNamesTestCase("empty exclusions", core.S[string]()),
	newGetInterfacesNamesTestCase("exclude invalid", core.S("invalid-interface")),
	newGetInterfacesNamesTestCase("exclude multiple", core.S("eth0", "eth1", "lo")),
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

	names, err := core.GetInterfacesNames(tc.except...)
	core.AssertNoError(t, err, "GetInterfacesNames")
	core.AssertNotNil(t, names, "interface names slice")

	// Check that excluded names are not in the result
	for _, excluded := range tc.except {
		core.AssertFalse(t, core.SliceContains(names, excluded), "contains interface %q", excluded)
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, name := range names {
		core.AssertFalse(t, seen[name], "duplicate interface %q", name)
		seen[name] = true
	}
}

func TestGetInterfacesNames(t *testing.T) {
	core.RunTestCases(t, getInterfacesNamesTestCases)

	// Additional test: verify exclusion actually works
	t.Run("exclusion removes interfaces", testExclusionRemovesInterfaces)
}

func testExclusionRemovesInterfaces(t *testing.T) {
	t.Helper()
	all, err := core.GetInterfacesNames()
	if err != nil || len(all) == 0 {
		t.Skip("No interfaces available for testing")
	}

	// Exclude the first interface
	excluded := all[0]
	filtered, err := core.GetInterfacesNames(excluded)
	core.AssertMustNoError(t, err, "GetInterfacesNames")

	// Should have one less interface
	core.AssertEqual(t, len(all)-1, len(filtered), "interface count")

	// Verify the excluded interface is not present
	core.AssertFalse(t, core.SliceContains(filtered, excluded), "contains interface %q", excluded)
}

// Benchmarks
func BenchmarkParseAddr(b *testing.B) {
	for b.Loop() {
		_, _ = core.ParseAddr("192.168.1.1")
	}
}

func BenchmarkParseNetIP(b *testing.B) {
	for b.Loop() {
		_, _ = core.ParseNetIP("192.168.1.1")
	}
}

func BenchmarkGetStringIPAddresses(b *testing.B) {
	for b.Loop() {
		_, _ = core.GetStringIPAddresses()
	}
}

func BenchmarkAddrFromNetIP(b *testing.B) {
	addr := &net.IPAddr{IP: net.ParseIP("192.168.1.1")}
	for b.Loop() {
		_, _ = core.AddrFromNetIP(addr)
	}
}
