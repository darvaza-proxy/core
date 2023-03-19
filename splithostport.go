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
	} else if host == "" {
		// :port
		host = "::" // use undetermined host
	} else if strings.ContainsRune(host, ':') {
		// unbracketed ipv6?
		if host, ok = validIP(hostport); ok {
			return host, "", nil
		}

		err = addrErr(hostport, "invalid address")
	}

	if err == nil {
		// successful split, but is it valid?

		if port != "" && !validPort(port) {
			// bad port
			err = addrErr(hostport, "invalid port")
		} else if s, ok := validIP(host); ok {
			// valid IP
			return s, port, nil
		} else if s, ok := validName(host); ok {
			// valid name
			return s, port, nil
		} else {
			// bad address
			err = addrErr(hostport, "invalid address")
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

func validIP(s string) (string, bool) {
	addr, err := ParseAddr(s)
	if err == nil {
		return addr.String(), true
	}
	return "", false
}

func validName(s string) (string, bool) {
	s, err := idna.Display.ToUnicode(s)
	if err == nil {
		return s, true
	}
	return "", false
}

func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
}
