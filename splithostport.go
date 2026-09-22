package core

import (
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

// MakeHostPort produces a validated host:port from an input string
// optionally using the given default port when the string doesn't
// specify one. The host comes back cleaned, an IPv6 address bracketed
// when a port follows it, and the port in canonical decimal form.
// Port 0 on the string input, in any spelling, isn't considered
// valid, and the error names the input as given.
//
// Examples:
//   - MakeHostPort("localhost", 8080) → "localhost:8080"
//   - MakeHostPort("localhost:9000", 8080) → "localhost:9000"
//   - MakeHostPort("192.168.1.1", 80) → "192.168.1.1:80"
//   - MakeHostPort("::1", 443) → "[::1]:443"
//   - MakeHostPort("example.com", 0) → "example.com"
//   - MakeHostPort("example.com:0080", 0) → "example.com:80"
//   - MakeHostPort("example.com:0", 80) → error (port 0 invalid)
//   - MakeHostPort("example.com:00", 80) → error (port 0 invalid)
func MakeHostPort(hostPort string, defaultPort uint16) (string, error) {
	host, port, err := SplitHostPort(hostPort)
	if err != nil {
		// bad input
		return "", err
	}

	// port is either valid or empty
	// host is either a valid hostname or a valid IP address

	if ip, _ := ParseAddr(host); ip.IsValid() {
		// IP address

		if port == "" && defaultPort == 0 {
			// portless IP
			return ip.String(), nil
		}

		host = ipForHostPort(ip)
	}

	s, ok := doMakeHostPort(host, port, defaultPort)
	if !ok {
		// port 0 in any spelling
		return "", addrErr(hostPort, "invalid port")
	}

	return s, nil
}

// doMakeHostPort joins a cleaned host and port, the port in canonical
// decimal form, and reports false for a port it refuses.
func doMakeHostPort(host, port string, defaultPort uint16) (string, bool) {
	if port == "" {
		if defaultPort == 0 {
			// portless hostname
			return host, true
		}
		return host + ":" + formatPort(defaultPort), true
	}

	n, err := parsePort(port)
	if err != nil || n == 0 {
		// bad port, or 0 in any spelling
		return "", false
	}

	return host + ":" + formatPort(n), true
}

// JoinHostPort is like the standard net.JoinHostPort, but
// it validates the host name and port, and returns it portless
// if the port argument is empty.
//
// Unlike net.JoinHostPort, this function:
//   - Validates the host and returns it cleaned: an IP in its
//     canonical text, a name in Unicode
//   - Validates the port is in range 0-65535 and joins it in
//     canonical decimal form
//   - Returns host without port if port is empty
//   - Properly handles IPv6 addresses with bracketing
//   - Supports international domain names
//   - Returns descriptive errors for invalid inputs
//
// Examples:
//   - JoinHostPort("localhost", "8080") → "localhost:8080"
//   - JoinHostPort("localhost", "") → "localhost"
//   - JoinHostPort("192.168.1.1", "80") → "192.168.1.1:80"
//   - JoinHostPort("::1", "443") → "[::1]:443"
//   - JoinHostPort("::1", "") → "::1"
//   - JoinHostPort("example.com", "0080") → "example.com:80"
//   - JoinHostPort("example.com", "0") → "example.com:0" (port 0 allowed)
//   - JoinHostPort("invalid host", "80") → error
//   - JoinHostPort("example.com", "99999") → error (port out of range)
func JoinHostPort(host, port string) (string, error) {
	ip, _ := ParseAddr(host)
	switch {
	case ip.IsValid():
		if port == "" {
			// portless IP ready
			return ip.String(), nil
		}

		host = ipForHostPort(ip)
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
	s, ok := canonicalPort(port)
	if !ok {
		// bad port
		return "", addrErr(host+":"+port, "invalid port")
	}

	return host + ":" + s, nil
}

// SplitHostPort is like net.SplitHostPort but doesn't fail if the
// port isn't part of the string and it validates it if present.
// SplitHostPort will also validate the host is a valid IP or name
//
// Unlike net.SplitHostPort, this function:
//   - Accepts hostport strings without port (returns empty port)
//   - Validates the host is a valid IP address or hostname, and returns
//     it cleaned: an IP in its canonical text, a name in Unicode
//   - Validates the port is in range 0-65535 and returns it in
//     canonical decimal form
//   - Properly handles IPv6 addresses with and without brackets
//   - Supports international domain names with punycode conversion
//   - Returns descriptive errors for invalid inputs
//
// Examples:
//   - SplitHostPort("localhost:8080") → ("localhost", "8080", nil)
//   - SplitHostPort("localhost:0080") → ("localhost", "80", nil)
//   - SplitHostPort("localhost") → ("localhost", "", nil)
//   - SplitHostPort("192.168.1.1:80") → ("192.168.1.1", "80", nil)
//   - SplitHostPort("[::1]:443") → ("::1", "443", nil)
//   - SplitHostPort("::1") → ("::1", "", nil)
//   - SplitHostPort("example.com") → ("example.com", "", nil)
//   - SplitHostPort("Hello.世界") → ("hello.世界", "", nil)
//   - SplitHostPort("invalid host") → error
//   - SplitHostPort("example.com:99999") → error (port out of range)
func SplitHostPort(hostPort string) (host, port string, err error) {
	host, port, err = splitHostPortUnsafe(hostPort)
	if err != nil {
		// failed split
		return "", "", err
	}

	if port != "" {
		s, ok := canonicalPort(port)
		if !ok {
			// bad port
			return "", "", addrErr(hostPort, "invalid port")
		}
		port = s
	}

	if s, ok := validIP(host); ok {
		// valid IP
		return s, port, nil
	}

	if s, ok := validName(host); ok {
		// valid name
		return s, port, nil
	}

	return "", "", addrErr(hostPort, "invalid address")
}

// SplitAddrPort splits a string containing an IP address and an optional port,
// and validates it. Returns the address as netip.Addr and port as uint16.
//
// This function:
//   - Accepts IP addresses with optional port numbers
//   - Validates the address is a valid IPv4 or IPv6 address
//   - Validates the port is in range 0-65535; the port is 0 when the
//     string carries none, as it is when the string says 0
//   - Properly handles IPv6 addresses with brackets
//   - Returns zero values and error for invalid inputs
//
// Examples:
//   - SplitAddrPort("192.168.1.1:8080") → (192.168.1.1, 8080, nil)
//   - SplitAddrPort("192.168.1.1") → (192.168.1.1, 0, nil)
//   - SplitAddrPort("[::1]:443") → (::1, 443, nil)
//   - SplitAddrPort("::1") → (::1, 0, nil)
//   - SplitAddrPort("127.0.0.1:80") → (127.0.0.1, 80, nil)
//   - SplitAddrPort("invalid") → error
//   - SplitAddrPort("192.168.1.1:99999") → error (port out of range)
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
	default:
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

func formatPort(n uint16) string {
	return strconv.FormatUint(uint64(n), 10)
}

// canonicalPort validates a port string and returns it in canonical
// decimal form, so a spelling such as "0080" comes back as "80".
func canonicalPort(s string) (string, bool) {
	n, err := parsePort(s)
	if err != nil {
		return "", false
	}
	return formatPort(n), true
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

func ipForHostPort(ip netip.Addr) string {
	if ip.Is6() {
		return "[" + ip.String() + "]"
	}

	return ip.String()
}

func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
}
