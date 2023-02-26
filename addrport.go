package core

import (
	"net"
	"net/netip"
)

// AddrPort attempts to extract a netip.AddrPort from an object
func AddrPort(v any) (netip.AddrPort, bool) {
	// via interfaces
	if p, ok := v.(interface {
		AddrPort() netip.AddrPort
	}); ok {
		return p.AddrPort(), true
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

	// known types
	return typeSpecificAddrPort(v)
}

func typeSpecificAddrPort(v any) (netip.AddrPort, bool) {
	switch addr := v.(type) {
	case *net.TCPAddr:
		if ip, ok := netip.AddrFromSlice(addr.IP); ok {
			return netip.AddrPortFrom(ip, uint16(addr.Port)), true
		}
	case *net.UDPAddr:
		if ip, ok := netip.AddrFromSlice(addr.IP); ok {
			return netip.AddrPortFrom(ip, uint16(addr.Port)), true
		}
	}

	return netip.AddrPort{}, false
}
