package core

import (
	"math"
	"net"
	"net/netip"
)

// AddrPort attempts to extract a netip.AddrPort from an object.
// It supports the following types:
//   - netip.AddrPort
//   - *netip.AddrPort (dereferenced)
//   - *net.TCPAddr (converted with IPv4 unmapping)
//   - *net.UDPAddr (converted with IPv4 unmapping)
//   - Types implementing AddrPort() netip.AddrPort method
//   - Types implementing Addr() net.Addr method (recursively processed)
//   - Types implementing RemoteAddr() net.Addr method (recursively processed)
//
// IPv4 addresses are properly unmapped, so 192.168.1.1:80 is returned
// instead of [::ffff:192.168.1.1]:80. A value that does not convert to a
// valid AddrPort returns false alongside the zero AddrPort.
func AddrPort(v any) (netip.AddrPort, bool) {
	// Known types first, and conclusively: a recognised type is not
	// offered to the interface arms below, which would answer the same
	// value differently.
	if addr, ok, known := typeSpecificAddrPort(v); known {
		return addr, ok
	}

	// via interfaces
	if p, ok := v.(interface {
		AddrPort() netip.AddrPort
	}); ok {
		return validAddrPort(p.AddrPort())
	}

	if p, ok := v.(interface {
		Addr() net.Addr
	}); ok {
		return AddrPort(p.Addr())
	}

	if p, ok := v.(interface {
		RemoteAddr() net.Addr
	}); ok {
		return AddrPort(p.RemoteAddr())
	}

	// sorry
	return netip.AddrPort{}, false
}

// validAddrPort passes a valid AddrPort through and discards an invalid one,
// so a rejected value is never returned alongside false.
func validAddrPort(ap netip.AddrPort) (netip.AddrPort, bool) {
	if !ap.IsValid() {
		return netip.AddrPort{}, false
	}
	return ap, true
}

// addrPortFromNetAddr creates an AddrPort from IP and port, properly unmapping IPv4.
// IPv4-mapped IPv6 addresses (::ffff:192.168.1.1) are converted to clean IPv4 addresses.
// Returns false if the IP is nil or invalid, or if the port falls outside the
// 16-bit range, which is rejected rather than truncated.
func addrPortFromNetAddr(ip net.IP, port int) (netip.AddrPort, bool) {
	p, ok := portFromInt(port)
	if !ok {
		return netip.AddrPort{}, false
	}

	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return netip.AddrPort{}, false
	}
	// Unmap to get clean IPv4 addresses
	return netip.AddrPortFrom(addr.Unmap(), p), true
}

// portFromInt narrows the int port a net.Addr carries to the 16-bit range
// a netip.AddrPort takes, rejecting a value outside it instead of
// truncating it. It is the only place that conversion is made.
func portFromInt(port int) (uint16, bool) {
	if port < 0 || port > math.MaxUint16 {
		return 0, false
	}
	return uint16(port), true
}

// typeSpecificAddrPort handles direct type conversions to AddrPort.
// It processes concrete types without interface checks. known reports whether
// v was one of those types, telling a value this rejected from one it never
// recognised.
func typeSpecificAddrPort(v any) (ap netip.AddrPort, ok, known bool) {
	switch addr := v.(type) {
	case netip.AddrPort:
		ap, ok = validAddrPort(addr)
	case *netip.AddrPort:
		ap, ok = validAddrPort(*addr)
	case *net.TCPAddr:
		ap, ok = addrPortFromNetAddr(addr.IP, addr.Port)
	case *net.UDPAddr:
		ap, ok = addrPortFromNetAddr(addr.IP, addr.Port)
	default:
		return netip.AddrPort{}, false, false
	}

	return ap, ok, true
}
