package core

import (
	"net"
	"net/netip"
	"testing"
)

func TestAsStringIPAddresses(t *testing.T) {
	addrs := S(
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("2001:db8::1"),
		netip.Addr{}, // Invalid address
		netip.MustParseAddr("10.0.0.1"),
	)

	result := asStringIPAddresses(addrs...)
	expected := S("192.168.1.1", "2001:db8::1", "10.0.0.1")

	AssertSliceEqual(t, expected, result, "addresses")
}

func TestAsNetIPAddresses(t *testing.T) {
	addrs := S(
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("2001:db8::1"),
		netip.Addr{},                              // Invalid address
		netip.MustParseAddr("::ffff:192.168.1.2"), // IPv4-mapped IPv6
	)

	result := asNetIPAddresses(addrs...)

	// Should have 3 valid addresses (invalid one skipped)
	AssertMustEqual(t, 3, len(result), "count")

	// Verify the IPv4-mapped address is unmapped
	lastAddr := result[2]
	AssertEqual(t, net.ParseIP("192.168.1.2"), lastAddr, "unmapped IPv4")
}

func TestAsNetIP(t *testing.T) {
	t.Run("IPv4", testAsNetIPv4)
	t.Run("IPv6", testAsNetIPv6)
}

func testAsNetIPv4(t *testing.T) {
	addr := netip.MustParseAddr("192.168.1.1")
	ip := asNetIP(addr)
	expected := net.ParseIP("192.168.1.1").To4()
	AssertEqual(t, expected, ip, "IPv4")
	// Should be 4 bytes for IPv4
	AssertEqual(t, 4, len(ip), "byte length")
}

func testAsNetIPv6(t *testing.T) {
	addr := netip.MustParseAddr("2001:db8::1")
	ip := asNetIP(addr)
	expected := net.ParseIP("2001:db8::1")
	AssertEqual(t, expected, ip, "IPv6")
	// Should be 16 bytes for IPv6
	AssertEqual(t, 16, len(ip), "byte length")
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

	// Should have 2 valid addresses, the TCPAddr and the nil IP skipped
	expected := S(
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("10.0.0.0"),
	)
	AssertSliceEqual(t, expected, result, "addresses")
}
