package core

import (
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

// SplitHostPort is like net.SplitHostPort but doesn't fail if the
// port isn't part of the string and it validates it if present.
// SplitHostPort will also validate the host is a valid IP or name
func SplitHostPort(hostport string) (host, port string, err error) {
	host, port, err = splitHostPortUnsafe(hostport)

	switch {
	case err != nil:
		// failed split
		return "", "", err
	case port != "" && !validPort(port):
		// bad port
		err = addrErr(hostport, "invalid port")
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

		err = addrErr(hostport, "invalid address")
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

func splitHostPortUnsafe(hostport string) (host, port string, err error) {
	var ok bool

	switch {
	case hostport == "":
		// empty
		err = addrErr(hostport, "empty address")
		return "", "", err
	case hostport[0] == '[':
		// [host]:port [host]
		return splitHostPortBracketed(hostport)
	case strings.Count(hostport, ":") > 1:
		// unbracketed IPv6
		return hostport, "", nil
	}

	host, port, ok = splitLastRune(':', hostport)
	switch {
	case !ok:
		// host without port
		host, port = hostport, ""
	case port == "":
		// host:
		err = addrErr(hostport, "missing port after ':'")
	case host == "":
		// :port
		host = "::" // use undetermined host
	}

	return host, port, err
}

func splitHostPortBracketed(hostport string) (host, port string, err error) {
	host, s, ok := splitLastRune(']', hostport[1:])
	switch {
	case !ok:
		// [host
		host = ""
		err = addrErr(hostport, "missing ']' in address")
	case s == "":
		// [host]
	case s[0] == ':':
		// [host]:...
		port = s[1:]
		if port == "" {
			// [host]:
			err = addrErr(hostport, "missing port after ':'")
		}
	default:
		// [host]...
		host = ""
		err = addrErr(hostport, "invalid character after ']'")
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
