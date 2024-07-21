package core

import (
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

// JoinHostPort is like the standard net.JoinHostPort, but
// it validates the host name and port, and returns it portless
// if the port argument is empty.
func JoinHostPort(host, port string) (string, error) {
	ip, _ := ParseAddr(host)
	switch {
	case ip.IsValid():
		switch {
		case port == "":
			// portless IP ready
			return ip.String(), nil
		case ip.Is6():
			// IPv6
			host = "[" + ip.String() + "]"
		default:
			// IPv4
			host = ip.String()
		}
	default:
		// not IP address
		s, ok := validName(host)
		switch {
		case !ok:
			// bad host name
			return "", addrErr(host, "invalid host")
		case port == "":
			// portless host
			return s, nil
		default:
			// good name
			host = s
		}
	}

	return doJoinHostPort(host, port)
}

func doJoinHostPort(host, port string) (string, error) {
	hostPort := host + ":" + port
	if !validPort(port) {
		// bad port
		return "", addrErr(hostPort, "invalid port")
	}

	return hostPort, nil
}

// SplitHostPort is like net.SplitHostPort but doesn't fail if the
// port isn't part of the string and it validates it if present.
// SplitHostPort will also validate the host is a valid IP or name
func SplitHostPort(hostPort string) (host, port string, err error) {
	host, port, err = splitHostPortUnsafe(hostPort)

	switch {
	case err != nil:
		// failed split
		return "", "", err
	case port != "" && !validPort(port):
		// bad port
		err = addrErr(hostPort, "invalid port")
		return "", "", err
	default:
		if s, ok := validIP(host); ok {
			// valid IP
			return s, port, nil
		}

		if s, ok := validName(host); ok {
			// valid name
			return s, port, nil
		}

		err = addrErr(hostPort, "invalid address")
		return "", "", err
	}
}

// SplitAddrPort splits a string containing an IP address and an optional port,
// and validates it.
func SplitAddrPort(addrPort string) (addr netip.Addr, port uint16, err error) {
	// split
	host, sPort, err := splitHostPortUnsafe(addrPort)
	if err != nil {
		// failed to split
		return netip.Addr{}, 0, err
	}

	// port
	if sPort != "" {
		port, err = parsePort(sPort)
		if err != nil {
			// bad port
			err = addrErr(addrPort, "invalid port")
			return netip.Addr{}, 0, err
		}
	}

	// addr
	addr, err = ParseAddr(host)
	if err != nil {
		// bad address
		err = addrErr(addrPort, "invalid address")
		return netip.Addr{}, 0, err
	}

	// success
	return addr, port, nil
}

func splitHostPortUnsafe(hostPort string) (host, port string, err error) {
	var ok bool

	switch {
	case hostPort == "":
		// empty
		err = addrErr(hostPort, "empty address")
		return "", "", err
	case hostPort[0] == '[':
		// [host]:port [host]
		return splitHostPortBracketed(hostPort)
	case strings.Count(hostPort, ":") > 1:
		// unbracketed IPv6
		return hostPort, "", nil
	}

	host, port, ok = splitLastRune(':', hostPort)
	switch {
	case !ok:
		// host without port
		host, port = hostPort, ""
	case port == "":
		// host:
		err = addrErr(hostPort, "missing port after ':'")
	case host == "":
		// :port
		host = "::" // use undetermined host
	}

	return host, port, err
}

func splitHostPortBracketed(hostPort string) (host, port string, err error) {
	host, s, ok := splitLastRune(']', hostPort[1:])
	switch {
	case !ok:
		// [host
		host = ""
		err = addrErr(hostPort, "missing ']' in address")
	case s == "":
		// [host]
	case s[0] == ':':
		// [host]:...
		port = s[1:]
		if port == "" {
			// [host]:
			err = addrErr(hostPort, "missing port after ':'")
		}
	default:
		// [host]...
		host = ""
		err = addrErr(hostPort, "invalid character after ']'")
	}

	return host, port, err
}

func splitLastRune(r rune, s string) (before, after string, found bool) {
	i := strings.LastIndexFunc(s, func(v rune) bool {
		return r == v
	})
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+1:], true
}

func parsePort(s string) (uint16, error) {
	u64, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(u64), nil
}

func validPort(s string) bool {
	_, err := parsePort(s)
	return err == nil
}

func validIP(s string) (string, bool) {
	addr, err := ParseAddr(s)
	if err == nil {
		return addr.String(), true
	}
	return "", false
}

var nameRE = regexp.MustCompile(`^(([\p{L}\p{M}\p{N}_%+-]+\.)+)?[\p{L}\p{M}\p{N}-]+$`)

func validName(s string) (string, bool) {
	if nameRE.MatchString(s) {
		s, err := idna.Display.ToUnicode(s)
		if err == nil {
			return s, true
		}
	}

	return "", false
}

func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
}
