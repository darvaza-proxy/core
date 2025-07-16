package core

import (
	"net"
	"net/netip"
	"reflect"
	"testing"
)

// Test cases for ParseAddr
type parseAddrTestCase struct {
	// netip.Addr struct (24 bytes) - result value
	want netip.Addr

	// String fields (16 bytes each) - alphabetically ordered
	input string
	name  string

	// Boolean field (1 byte) - flags
	wantErr bool
}

var parseAddrTestCases = []parseAddrTestCase{
	{
		name:    "IPv4 address",
		input:   "192.168.1.1",
		want:    netip.MustParseAddr("192.168.1.1"),
		wantErr: false,
	},
	{
		name:    "IPv6 address",
		input:   "2001:db8::1",
		want:    netip.MustParseAddr("2001:db8::1"),
		wantErr: false,
	},
	{
		name:    "IPv4 unspecified shorthand",
		input:   "0",
		want:    netip.IPv4Unspecified(),
		wantErr: false,
	},
	{
		name:    "IPv6 unspecified",
		input:   "::",
		want:    netip.IPv6Unspecified(),
		wantErr: false,
	},
	{
		name:    "IPv4 loopback",
		input:   "127.0.0.1",
		want:    netip.MustParseAddr("127.0.0.1"),
		wantErr: false,
	},
	{
		name:    "IPv6 loopback",
		input:   "::1",
		want:    netip.MustParseAddr("::1"),
		wantErr: false,
	},
	{
		name:    "Invalid address",
		input:   "not-an-ip",
		want:    netip.Addr{},
		wantErr: true,
	},
	{
		name:    "Empty string",
		input:   "",
		want:    netip.Addr{},
		wantErr: true,
	},
	{
		name:    "IPv4 with port",
		input:   "192.168.1.1:8080",
		want:    netip.Addr{},
		wantErr: true,
	},
	{
		name:    "IPv6 with zone",
		input:   "fe80::1%eth0",
		want:    netip.MustParseAddr("fe80::1%eth0"),
		wantErr: false,
	},
}

func (tc parseAddrTestCase) test(t *testing.T) {
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
	for _, tc := range parseAddrTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for ParseNetIP
type parseNetIPTestCase struct {
	name    string
	input   string
	want    net.IP
	wantErr bool
}

var parseNetIPTestCases = []parseNetIPTestCase{
	{
		name:    "IPv4 address",
		input:   "192.168.1.1",
		want:    net.ParseIP("192.168.1.1"),
		wantErr: false,
	},
	{
		name:    "IPv6 address",
		input:   "2001:db8::1",
		want:    net.ParseIP("2001:db8::1"),
		wantErr: false,
	},
	{
		name:    "IPv4 unspecified shorthand",
		input:   "0",
		want:    net.IPv4zero,
		wantErr: false,
	},
	{
		name:    "IPv6 unspecified",
		input:   "::",
		want:    net.IPv6zero,
		wantErr: false,
	},
	{
		name:    "Invalid address",
		input:   "invalid",
		want:    nil,
		wantErr: true,
	},
	{
		name:    "Empty string",
		input:   "",
		want:    nil,
		wantErr: true,
	},
}

func (tc parseNetIPTestCase) test(t *testing.T) {
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
	for _, tc := range parseNetIPTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test cases for AddrFromNetIP
type addrFromNetIPTestCase struct {
	// netip.Addr struct (24 bytes) - result value
	want netip.Addr

	// Interface field (16 bytes) - input value
	input net.Addr

	// String field (16 bytes) - test name
	name string

	// Boolean field (1 byte) - success flag
	ok bool
}

var addrFromNetIPTestCases = []addrFromNetIPTestCase{
	{
		name: "IPAddr IPv4",
		input: &net.IPAddr{
			IP: net.ParseIP("192.168.1.1"),
		},
		want: netip.MustParseAddr("192.168.1.1"),
		ok:   true,
	},
	{
		name: "IPAddr IPv6",
		input: &net.IPAddr{
			IP: net.ParseIP("2001:db8::1"),
		},
		want: netip.MustParseAddr("2001:db8::1"),
		ok:   true,
	},
	{
		name: "IPNet IPv4",
		input: &net.IPNet{
			IP:   net.ParseIP("10.0.0.0"),
			Mask: net.CIDRMask(24, 32),
		},
		want: netip.MustParseAddr("10.0.0.0"),
		ok:   true,
	},
	{
		name: "IPNet IPv6",
		input: &net.IPNet{
			IP:   net.ParseIP("2001:db8::"),
			Mask: net.CIDRMask(64, 128),
		},
		want: netip.MustParseAddr("2001:db8::"),
		ok:   true,
	},
	{
		name:  "TCPAddr",
		input: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		want:  netip.Addr{},
		ok:    false,
	},
	{
		name:  "UDPAddr",
		input: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
		want:  netip.Addr{},
		ok:    false,
	},
	{
		name:  "nil input",
		input: nil,
		want:  netip.Addr{},
		ok:    false,
	},
	{
		name: "IPAddr with nil IP",
		input: &net.IPAddr{
			IP: nil,
		},
		want: netip.Addr{},
		ok:   false,
	},
	{
		name: "IPNet with nil IP",
		input: &net.IPNet{
			IP:   nil,
			Mask: net.CIDRMask(24, 32),
		},
		want: netip.Addr{},
		ok:   false,
	},
}

func (tc addrFromNetIPTestCase) test(t *testing.T) {
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
	for _, tc := range addrFromNetIPTestCases {
		t.Run(tc.name, tc.test)
	}
}

// Test GetStringIPAddresses
//
//revive:disable-next-line:cognitive-complexity
func TestGetStringIPAddresses(t *testing.T) {
	// Since we can't easily mock net.InterfaceByName and net.InterfaceAddrs,
	// we'll test with the real interfaces but handle the case where there are none

	t.Run("all interfaces", func(t *testing.T) {
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
	})

	t.Run("specific interface", func(t *testing.T) {
		// Try with a non-existent interface
		_, err := GetStringIPAddresses("invalid-interface-name")
		if err == nil {
			t.Error("Expected error for invalid interface")
		}
	})

	t.Run("loopback interface", testLoopbackInterface)
}

func testLoopbackInterface(t *testing.T) {
	// Most systems have a loopback interface
	loopbackNames := []string{"lo", "lo0", "Loopback Pseudo-Interface 1"}

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

// Test GetNetIPAddresses
//
//revive:disable-next-line:cognitive-complexity
func TestGetNetIPAddresses(t *testing.T) {
	t.Run("all interfaces", func(t *testing.T) {
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
	})

	t.Run("invalid interface", func(t *testing.T) {
		_, err := GetNetIPAddresses("invalid-interface-name")
		if err == nil {
			t.Error("Expected error for invalid interface")
		}
	})
}

// Test GetIPAddresses
//
//revive:disable-next-line:cognitive-complexity
func TestGetIPAddresses(t *testing.T) {
	t.Run("all interfaces", func(t *testing.T) {
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
	})

	t.Run("multiple interfaces with error", func(t *testing.T) {
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
	})
}

// Test GetInterfacesNames
type getInterfacesNamesTestCase struct {
	name   string
	except []string
}

var getInterfacesNamesTestCases = []getInterfacesNamesTestCase{
	{
		name:   "no exclusions",
		except: nil,
	},
	{
		name:   "empty exclusions",
		except: []string{},
	},
	{
		name:   "exclude invalid",
		except: []string{"invalid-interface"},
	},
	{
		name:   "exclude multiple",
		except: []string{"eth0", "eth1", "lo"},
	},
}

//revive:disable-next-line:cognitive-complexity
func (tc getInterfacesNamesTestCase) test(t *testing.T) {
	t.Helper()

	names, err := GetInterfacesNames(tc.except...)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if names == nil {
		t.Error("Expected non-nil slice")
		return
	}

	// Check that excluded names are not in the result
	for _, excluded := range tc.except {
		for _, name := range names {
			if name == excluded {
				t.Errorf("Found excluded interface %s in result", excluded)
			}
		}
	}

	// Verify no duplicates
	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			t.Errorf("Duplicate interface name: %s", name)
		}
		seen[name] = true
	}
}

//revive:disable-next-line:cognitive-complexity
func TestGetInterfacesNames(t *testing.T) {
	for _, tc := range getInterfacesNamesTestCases {
		t.Run(tc.name, tc.test)
	}

	// Additional test: verify exclusion actually works
	t.Run("exclusion removes interfaces", func(t *testing.T) {
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
	})
}

// Test internal helper functions
func TestAsStringIPAddresses(t *testing.T) {
	addrs := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("2001:db8::1"),
		{}, // Invalid address
		netip.MustParseAddr("10.0.0.1"),
	}

	result := asStringIPAddresses(addrs...)
	expected := []string{"192.168.1.1", "2001:db8::1", "10.0.0.1"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestAsNetIPAddresses(t *testing.T) {
	addrs := []netip.Addr{
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("2001:db8::1"),
		{}, // Invalid address
		netip.MustParseAddr("::ffff:192.168.1.2"), // IPv4-mapped IPv6
	}

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

//revive:disable-next-line:cognitive-complexity
func TestAsNetIP(t *testing.T) {
	t.Run("IPv4", func(t *testing.T) {
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
	})

	t.Run("IPv6", func(t *testing.T) {
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
	})
}

func TestAppendNetIPAsIP(t *testing.T) {
	var out []netip.Addr

	addrs := []net.Addr{
		&net.IPAddr{IP: net.ParseIP("192.168.1.1")},
		&net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)},
		&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, // Should be skipped
		&net.IPAddr{IP: nil}, // Should be skipped
	}

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
