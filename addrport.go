package core

import (
	"net"
	"net/netip"
)

// AddrPort attempts to extract a netip.AddrPort from an object.
// It supports the following types:
//   - netip.AddrPort (returned as-is)
//   - *netip.AddrPort (dereferenced)
//   - *net.TCPAddr (converted with IPv4 unmapping)
//   - *net.UDPAddr (converted with IPv4 unmapping)
//   - Types implementing AddrPort() netip.AddrPort method (if result is valid)
//   - Types implementing Addr() net.Addr method (recursively processed)
//   - Types implementing RemoteAddr() net.Addr method (recursively processed)
//
// IPv4 addresses are properly unmapped, so 192.168.1.1:80 is returned
// instead of [::ffff:192.168.1.1]:80. Invalid AddrPort values return false.
func AddrPort(v any) (netip.AddrPort, bool) {
	// known types first
	if addr, ok := typeSpecificAddrPort(v); ok {
		return addr, ok
	}

	// via interfaces
	if p, ok := v.(interface {
		AddrPort() netip.AddrPort
	}); ok {
		ap := p.AddrPort()
		// Only return true if the AddrPort is valid
		return ap, ap.IsValid()
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

// addrPortFromNetAddr creates an AddrPort from IP and port, properly unmapping IPv4.
// IPv4-mapped IPv6 addresses (::ffff:192.168.1.1) are converted to clean IPv4 addresses.
// Returns false if the IP is nil or invalid.
func addrPortFromNetAddr(ip net.IP, port int) (netip.AddrPort, bool) {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return netip.AddrPort{}, false
	}
	// Unmap to get clean IPv4 addresses
	return netip.AddrPortFrom(addr.Unmap(), uint16(port)), true
}

// typeSpecificAddrPort handles direct type conversions to AddrPort.
// It processes concrete types without interface checks.
func typeSpecificAddrPort(v any) (netip.AddrPort, bool) {
	switch addr := v.(type) {
	case netip.AddrPort:
		return addr, true
	case *netip.AddrPort:
		return *addr, true
	case *net.TCPAddr:
		return addrPortFromNetAddr(addr.IP, addr.Port)
	case *net.UDPAddr:
		return addrPortFromNetAddr(addr.IP, addr.Port)
	}

	return netip.AddrPort{}, false
}
