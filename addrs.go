package core

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
)

// GetStringIPAddresses returns a list of text IP addresses bound
// to the given interfaces or all if none are given
func GetStringIPAddresses(ifaces ...string) ([]string, error) {
	addrs, err := GetIPAddresses(ifaces...)
	out := asStringIPAddresses(addrs...)

	return out, err
}

func asStringIPAddresses(addrs ...netip.Addr) []string {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if addr.IsValid() {
			s := addr.String()
			out = append(out, s)
		}
	}
	return out
}

// GetNetIPAddresses returns a list of net.IP addresses bound to
// the given interfaces or all if none are given
func GetNetIPAddresses(ifaces ...string) ([]net.IP, error) {
	addrs, err := GetIPAddresses(ifaces...)
	out := asNetIPAddresses(addrs...)
	return out, err
}

func asNetIPAddresses(addrs ...netip.Addr) []net.IP {
	out := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		if addr.IsValid() {
			out = append(out, asNetIP(addr.Unmap()))
		}
	}

	return out
}

func asNetIP(addr netip.Addr) net.IP {
	if addr.Is4() {
		a4 := addr.As4()
		return a4[:]
	}

	a16 := addr.As16()
	return a16[:]
}

// GetIPAddresses returns a list of netip.Addr bound to the given
// interfaces or all if none are given
func GetIPAddresses(ifaces ...string) ([]netip.Addr, error) {
	var out []netip.Addr

	if len(ifaces) == 0 {
		// all addresses
		addrs, err := net.InterfaceAddrs()
		out = appendNetIPAsIP(out, addrs...)

		return out, err
	}

	// only given
	for _, name := range ifaces {
		ifi, err := net.InterfaceByName(name)
		if err != nil {
			return out, err
		}

		addrs, err := ifi.Addrs()
		if err != nil {
			return out, err
		}

		out = appendNetIPAsIP(out, addrs...)
	}

	return out, nil
}

func appendNetIPAsIP(out []netip.Addr, addrs ...net.Addr) []netip.Addr {
	for _, ip := range addrs {
		addr, ok := AddrFromNetIP(ip)
		if ok && addr.IsValid() {
			out = append(out, addr.Unmap())
		}
	}

	return out
}

// AddrFromNetIP attempts to convert a net.Addr into a netip.Addr
func AddrFromNetIP(addr net.Addr) (netip.Addr, bool) {
	var s []byte

	switch v := addr.(type) {
	case *net.IPAddr:
		s = v.IP
	case *net.IPNet:
		s = v.IP
	}

	return netip.AddrFromSlice(s)
}

// GetInterfacesNames returns the list of interfaces,
// considering an optional exclusion list
func GetInterfacesNames(except ...string) ([]string, error) {
	s, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(s))

	for _, ifi := range s {
		if s := ifi.Name; s != "" {
			out = append(out, s)
		}
	}

	if len(except) > 0 {
		out = SliceMinus(out, except)
	}
	return out, nil
}

// ParseAddr turns a string into netip.Addr
func ParseAddr(s string) (addr netip.Addr, err error) {
	switch s {
	case "0":
		addr = netip.IPv4Unspecified()
	case "::":
		addr = netip.IPv6Unspecified()
	default:
		addr, err = netip.ParseAddr(s)
		if err != nil {
			return addr, err
		}
	}

	return addr, nil
}

// ParseNetIP turns a string into a net.IP
func ParseNetIP(s string) (ip net.IP, err error) {
	addr, err := ParseAddr(s)
	if err != nil {
		return nil, err
	}

	return asNetIP(addr.Unmap()), nil
}

// ParsePort validates a string as a TCP port
func ParsePort(s string) (port uint16, err error) {
	if s == "" {
		return 0, nil
	}

	prt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	outOfRangeErr := fmt.Errorf("value out of range for a port")
	max := int64(^uint16(0))

	if 0 > prt {
		if prt > -max {
			return uint16(-prt), nil
		}
		return 0, outOfRangeErr
	}
	if prt <= max {
		return uint16(prt), nil
	}
	return 0, outOfRangeErr
}
