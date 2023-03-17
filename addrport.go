package core

import (
	"net"
	"net/netip"
)

// SplitHostPort is like net.SplitHostPort but doesn't fail if the
// port isn't part of the string
func SplitHostPort(hostport string) (host, port string, err error) {
	h, p, err := net.SplitHostPort(hostport)
	if err == nil {
		// good pair
		return h, p, nil
	}

	if addrErr, ok := err.(*net.AddrError); ok {
		if addrErr.Err == "missing port in address" {
			// didn't have a port. let's add one and validate the host
			// part again
			h, _, err = SplitHostPort(net.JoinHostPort(hostport, "0"))
			return h, "", err
		}
	}

	return "", "", err
}

// AddrPort attempts to extract a netip.AddrPort from an object
func AddrPort(v any) (netip.AddrPort, bool) {
	// known types first
	if addr, ok := typeSpecificAddrPort(v); ok {
		return addr, ok
	}

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

	// sorry
	return netip.AddrPort{}, false
}

func typeSpecificAddrPort(v any) (netip.AddrPort, bool) {
	switch addr := v.(type) {
	case netip.AddrPort:
		return addr, true
	case *netip.AddrPort:
		return *addr, true
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
