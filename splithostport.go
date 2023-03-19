package core

import (
	"net"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

// revive:disable:cognitive-complexity
// revive:disable:cyclomatic

// SplitHostPort is like net.SplitHostPort but doesn't fail if the
// port isn't part of the string and it validates it if present.
// SplitHostPort will also validate the host is a valid IP or name
func SplitHostPort(hostport string) (host, port string, err error) {
	// revive:enable:cognitive-complexity
	// revive:enable:cyclomatic
	var ok bool

	if hostport == "" {
		// empty
		err = addrErr(hostport, "empty address")
	} else if hostport[0] == '[' {
		// [host]:port [host]
		host, port, err = splitHostPortBracketed(hostport)
	} else if host, port, ok = splitLastRune(':', hostport); !ok {
		// host without port
		host, port = hostport, ""
	} else if port == "" {
		// host:
		err = addrErr(hostport, "missing port after ':'")
	} else if strings.ContainsRune(host, ':') {
		// unbracketed ipv6?
		if validIPv6(hostport) {
			return hostport, "", nil
		}

		err = addrErr(hostport, "invalid address")
	}

	if err == nil {
		// successful split, but is it valid?

		switch {
		case port != "" && !validPort(port):
			err = addrErr(hostport, "invalid port")
		case !validIP(host) && !validName(host):
			err = addrErr(hostport, "invalid address")
		default:
			// all good
			return host, port, nil
		}
	}

	return "", "", err
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

func validPort(s string) bool {
	_, err := strconv.ParseUint(s, 10, 16)
	return err == nil
}

func validIP(s string) bool {
	_, err := ParseAddr(s)
	return err == nil
}

func validIPv6(s string) bool {
	addr, err := ParseAddr(s)
	return err == nil && addr.Is6()
}

func validName(s string) bool {
	_, err := idna.ToUnicode(s)
	return err == nil
}

func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
}
