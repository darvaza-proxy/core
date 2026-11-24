package core

import (
	"net"
	"net/netip"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

const hostLabelChars = "abcdefghijklmnopqrstuvwxyz0123456789_-"

// MakeHostPort turns a configured address into one to advertise,
// filling in the default port when the input names none. The other side
// splits it with SplitHostPort. The host comes back cleaned, an IPv6
// address bracketed when a port follows it, and the port in canonical
// decimal form. Port 0 in the input, in any spelling, leaves the port
// to be picked when a listener binds, a value MakeHostPort has no way
// to fill, so the input is refused, and the error names it as given.
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
	// host is either a valid hostname or a valid IP address, cleaned,
	// or empty for a port alone

	s, ok := doMakeHostPort(host, port, defaultPort)
	if !ok {
		// port 0 in any spelling
		return "", addrErr(hostPort, "invalid port")
	}

	return s, nil
}

// doMakeHostPort joins a cleaned host without brackets, or an empty one, and
// its port, the port in canonical decimal form, using defaultPort when
// port is empty, and returns the host alone when both are unset. It
// reports false for a port that does not parse, or parses to 0.
func doMakeHostPort(host, port string, defaultPort uint16) (string, bool) {
	switch {
	case port != "":
		n, err := parsePort(port)
		if err != nil || n == 0 {
			// bad port, or 0 in any spelling
			return "", false
		}
		port = formatPort(n)
	case defaultPort != 0:
		// default port
		port = formatPort(defaultPort)
	default:
		// neither port nor default, portless
	}

	return joinHostPortUnsafe(host, port), true
}

// JoinHostPort is like the standard net.JoinHostPort, but
// it validates the host name and port, and returns it portless
// if the port argument is empty. SplitHostPort splits the result
// back into its host and port.
//
// Unlike net.JoinHostPort, this function:
//   - Validates the host and returns it cleaned: an IP in its
//     canonical text, and a name in Unicode, case-folded as UTS 46
//     maps it, whose ASCII form is an RFC 1123 host name, '_' allowed,
//     within the lengths of RFC 1035. A name the mapping turns into an
//     IP address comes back as that address, and any other name whose
//     last label is a number is refused
//   - Validates the port is in range 0-65535 and joins it in
//     canonical decimal form
//   - Returns host without port if port is empty
//   - Returns ":port" if host is empty and port is not
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
//   - JoinHostPort("", "8080") → ":8080"
//   - JoinHostPort("invalid host", "80") → error
//   - JoinHostPort("example.com", "99999") → error (port out of range)
func JoinHostPort(host, port string) (string, error) {
	if host == "" && port != "" {
		// port alone
		return doJoinHostPort("", port)
	}

	s, ok := validHost(host)
	switch {
	case !ok:
		// bad host
		return "", addrErr(host, "invalid host")
	case port == "":
		// portless host
		return s, nil
	default:
		// good host
		return doJoinHostPort(s, port)
	}
}

// doJoinHostPort joins a cleaned host without brackets, or an empty one, and
// a port, the port in canonical decimal form. It refuses a port that
// does not parse, and the error names the host and port joined as
// passed.
func doJoinHostPort(host, port string) (string, error) {
	s, ok := canonicalPort(port)
	if !ok {
		// bad port
		return "", addrErr(joinHostPortUnsafe(host, port), "invalid port")
	}

	return joinHostPortUnsafe(host, s), nil
}

// SplitHostPort is like net.SplitHostPort but doesn't fail if the
// port isn't part of the string and it validates it if present.
// SplitHostPort will also validate the host is a valid IP or name.
//
// Unlike net.SplitHostPort, this function:
//   - Accepts host:port strings without port (returns empty port)
//   - Accepts a port alone, ":port" (returns empty host)
//   - Validates the host is a valid IP address or hostname, and returns
//     it cleaned: an IP in its canonical text, and a name in Unicode,
//     case-folded as UTS 46 maps it, whose ASCII form is an RFC 1123
//     host name, '_' allowed, within the lengths of RFC 1035. A name
//     the mapping turns into an IP address comes back as that address,
//     and any other name whose last label is a number is refused
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
//   - SplitHostPort(":8080") → ("", "8080", nil)
//   - SplitHostPort("invalid host") → error
//   - SplitHostPort("example.com:99999") → error (port out of range)
func SplitHostPort(hostPort string) (host, port string, err error) {
	_, _, host, port, err = doSplitHostPort(hostPort)
	return host, port, err
}

// SplitHostPortASCII splits and validates as SplitHostPort does, and
// returns a name in its ASCII form, a label outside ASCII in punycode.
// An IP address comes back in its canonical text, as from SplitHostPort.
//
// Examples:
//   - SplitHostPortASCII("Hello.世界:443") → ("hello.xn--rhqv96g", "443", nil)
//   - SplitHostPortASCII("example.com") → ("example.com", "", nil)
//   - SplitHostPortASCII("[::1]:443") → ("::1", "443", nil)
//   - SplitHostPortASCII("invalid host") → error
func SplitHostPortASCII(hostPort string) (host, port string, err error) {
	_, host, _, port, err = doSplitHostPort(hostPort)
	return host, port, err
}

// SplitAddrPortASCII splits and validates as SplitHostPort does, and
// returns the host in ASCII, and as an IP address where it is one. A
// host that is an IP address, or that UTS 46 maps into one, comes back
// as addr, as ParseAddr reads it, and as host in addr's text. Any other
// host is a name, and comes back as host in its ASCII form, a label
// outside ASCII in punycode, with the zero netip.Addr. A port alone
// comes back with neither. The port is in canonical decimal form, or
// empty where the input carries none.
//
// Examples:
//   - SplitAddrPortASCII("[::1]:443") → (::1, "::1", "443", nil)
//   - SplitAddrPortASCII("192.168.1.1") → (192.168.1.1, "192.168.1.1", "", nil)
//   - SplitAddrPortASCII("Hello.世界:443") → (netip.Addr{}, "hello.xn--rhqv96g", "443", nil)
//   - SplitAddrPortASCII(":8080") → (netip.Addr{}, "", "8080", nil)
//   - SplitAddrPortASCII("invalid host") → error
//
//revive:disable-next-line:function-result-limit // the address and the host, beside the port
func SplitAddrPortASCII(addrPort string) (addr netip.Addr, host, port string, err error) {
	addr, host, _, port, err = doSplitHostPort(addrPort)
	return addr, host, port, err
}

// doSplitHostPort splits and validates a host and port for
// SplitHostPort, SplitHostPortASCII and SplitAddrPortASCII. It returns
// the host as validHostForms does, nothing for a port alone, and the
// port in canonical decimal form.
//
//revive:disable-next-line:function-result-limit // the host in its forms, beside the port
func doSplitHostPort(hostPort string) (addr netip.Addr, ascii, unicode, port string, err error) {
	var ok bool

	// a failed split returns neither host nor port
	host, port, err := splitHostPortUnsafe(hostPort)
	if port != "" {
		// canonicalPort returns no port for a bad one
		port, ok = canonicalPort(port)
		if !ok {
			// bad port, leaving no host to check
			host, err = "", addrErr(hostPort, "invalid port")
		}
	}

	if host != "" {
		// validHostForms returns nothing for a bad host
		addr, ascii, unicode, ok = validHostForms(host)
		if !ok {
			// bad host
			port, err = "", addrErr(hostPort, "invalid address")
		}
	}

	return addr, ascii, unicode, port, err
}

// SplitAddrPort splits a string containing an IP address and an optional port,
// and validates it. Returns the address as netip.Addr and port as uint16.
//
// This function:
//   - Accepts IP addresses with optional port numbers
//   - Validates the address is a valid IPv4 or IPv6 address, or a host
//     UTS 46 maps into one, which comes back as that address
//   - Validates the port is in range 0-65535; the port is 0 when the
//     string carries none, as it is when the string says 0
//   - Accepts a port alone, returning the zero netip.Addr with it
//   - Properly handles IPv6 addresses with brackets
//   - Returns zero values and error for invalid inputs
//
// Examples:
//   - SplitAddrPort("192.168.1.1:8080") → (192.168.1.1, 8080, nil)
//   - SplitAddrPort("192.168.1.1") → (192.168.1.1, 0, nil)
//   - SplitAddrPort("[::1]:443") → (::1, 443, nil)
//   - SplitAddrPort("::1") → (::1, 0, nil)
//   - SplitAddrPort("127.0.0.1:80") → (127.0.0.1, 80, nil)
//   - SplitAddrPort(":8080") → (netip.Addr{}, 8080, nil)
//   - SplitAddrPort("invalid") → error
//   - SplitAddrPort("192.168.1.1:99999") → error (port out of range)
func SplitAddrPort(addrPort string) (addr netip.Addr, port uint16, err error) {
	var ok bool

	// a failed split returns neither host nor port
	host, sPort, err := splitHostPortUnsafe(addrPort)
	if sPort != "" {
		// parsePort returns 0 for a bad port
		port, err = parsePort(sPort)
		if err != nil {
			// bad port, leaving no address to check
			host, err = "", addrErr(addrPort, "invalid port")
		}
	}

	if host != "" {
		// validAddr returns the zero netip.Addr for a bad address
		addr, ok = validAddr(host)
		if !ok {
			// bad address
			port, err = 0, addrErr(addrPort, "invalid address")
		}
	}

	return addr, port, err
}

// splitHostPortUnsafe splits a host and port without validating either,
// brackets removed. On failure it returns the error alone.
func splitHostPortUnsafe(hostPort string) (host, port string, err error) {
	var ok bool

	switch {
	case hostPort == "":
		// empty
		err = addrErr(hostPort, "empty address")
		return host, port, err
	case hostPort[0] == '[':
		// [host]:port [host]
		return splitHostPortBracketed(hostPort)
	case strings.Count(hostPort, ":") > 1:
		// IPv6 without brackets
		return hostPort, "", nil
	default:
		// at most one ':', split below
	}

	host, port, ok = splitLastRune(':', hostPort)
	switch {
	case !ok:
		// host without port, as split
	case port == "":
		// host:
		host, err = "", addrErr(hostPort, "missing port after ':'")
	case host == "":
		// :port, no host
	default:
		// host:port, as split
	}

	return host, port, err
}

func splitHostPortBracketed(hostPort string) (host, port string, err error) {
	host, s, ok := splitLastRune(']', hostPort[1:])
	switch {
	case !ok:
		// [host
		host, err = "", addrErr(hostPort, "missing ']' in address")
	case host == "":
		// []...
		err = addrErr(hostPort, "missing address in brackets")
	case s == "":
		// [host]
	case s[0] == ':':
		// [host]:...
		port = s[1:]
		if port == "" {
			// [host]:
			host, err = "", addrErr(hostPort, "missing port after ':'")
		}
	default:
		// [host]...
		host, err = "", addrErr(hostPort, "invalid character after ']'")
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

// validHost validates a host as an IP address or a name, and returns it
// cleaned: an IP address in its canonical text, a name mapped through
// UTS 46 and in Unicode. A name the mapping turns into an IP address
// comes back as that address.
func validHost(s string) (string, bool) {
	_, _, unicode, ok := validHostForms(s)
	return unicode, ok
}

// validHostForms validates a host as an IP address or a name, and
// returns it in its forms. Where the host is an IP address, or UTS 46
// maps it into one, it returns the address, and its canonical text as
// both ascii and unicode. A name comes back with the zero netip.Addr,
// in the ASCII form the mapping gives, and decoded into Unicode, which
// only a name with a label in punycode needs. A host it refuses comes
// back as nothing but false.
//
//revive:disable-next-line:function-result-limit // the host in its forms
func validHostForms(s string) (addr netip.Addr, ascii, unicode string, ok bool) {
	addr, ascii, ok = mapHost(s)
	switch {
	case !ok:
		// refused by idna, nothing to return
	case addr.IsValid():
		// IP address
		ascii = addr.String()
		unicode = ascii
	case !validHostName(ascii):
		// not a host name
		ascii, ok = "", false
	case !hasALabel(ascii):
		// name in plain ASCII, its own Unicode form
		unicode = ascii
	default:
		// international name
		var err error
		unicode, err = nameProfile.ToUnicode(ascii)
		if err != nil {
			// idna refuses, once decoded, some names ToASCII
			// accepts, the bidi rule among others
			ascii, unicode, ok = "", "", false
		}
	}

	return addr, ascii, unicode, ok
}

// validAddr parses a host as an IP address, or as the IP address
// UTS 46 maps it into.
func validAddr(s string) (netip.Addr, bool) {
	addr, _, _ := mapHost(s)
	return addr, addr.IsValid()
}

// mapHost reads a host as an IP address, or maps it through UTS 46 and
// reads the result as one. It returns the address where either is one,
// and otherwise the ASCII form of the name, not yet checked by
// validHostName. It reports false where idna refuses the host.
func mapHost(s string) (netip.Addr, string, bool) {
	if addr, err := ParseAddr(s); err == nil {
		// IP address
		return addr, "", true
	}

	ascii, err := nameProfile.ToASCII(s)
	switch {
	case err != nil:
		// refused by idna
		return netip.Addr{}, "", false
	case ascii != s:
		// mapped, maybe into an IP address
		if addr, err := ParseAddr(ascii); err == nil {
			return addr, "", true
		}
	default:
		// unchanged, and already read as no IP address
	}

	// name
	return netip.Addr{}, ascii, true
}

// nameProfile maps a name through UTS 46 for lookup, with the STD3 and
// hyphen checks off, as validHostName checks characters and hyphens
// instead.
var nameProfile = idna.New(
	idna.MapForLookup(),
	idna.BidiRule(),
	idna.StrictDomainName(false),
	idna.CheckHyphens(false),
)

// validHostName reports whether an ASCII name is a host name: at most
// 253 octets, as RFC 1035 allows, labels validHostLabel accepts, and a
// last label that is not a number, as no top-level domain is.
func validHostName(s string) bool {
	if len(s) > 253 {
		// too long
		return false
	}

	for label := range strings.SplitSeq(s, ".") {
		if !validHostLabel(label) {
			// bad label
			return false
		}
	}
	return !isNumericLabel(lastLabel(s))
}

// validHostLabel reports whether an ASCII label is an RFC 1123 host name
// label, '_' allowed: 1 to 63 octets of lower-case letters, digits, '_'
// and '-', neither starting nor ending with '-', nor decoding, as an
// A-label, into a label that does.
func validHostLabel(s string) bool {
	switch {
	case s == "", len(s) > 63:
		// empty, or too long
		return false
	case s[0] == '-', s[len(s)-1] == '-':
		// hyphen at an end
		return false
	case strings.Trim(s, hostLabelChars) != "":
		// bad character
		return false
	default:
		return validALabelHyphens(s)
	}
}

// hasALabel reports whether an ASCII name has a label in punycode, an
// A-label, one starting "xn--".
func hasALabel(s string) bool {
	return strings.HasPrefix(s, "xn--") || strings.Contains(s, ".xn--")
}

// validALabelHyphens reports whether a label starting "xn--" decodes
// into one neither starting nor ending with '-', which its ASCII form
// does not show. It decodes only an A-label whose ASCII characters,
// those before its last '-', start or end with one. Any other label
// passes.
func validALabelHyphens(s string) bool {
	enc, ok := strings.CutPrefix(s, "xn--")
	if !ok {
		// not an A-label
		return true
	}

	basic, _, _ := splitLastRune('-', enc)
	if !strings.HasPrefix(basic, "-") && !strings.HasSuffix(basic, "-") {
		// no hyphen at either end of its ASCII characters
		return true
	}

	// nameProfile.ToASCII has already refused a label that fails to
	// decode
	u := Must(idna.Punycode.ToUnicode(s))
	return !strings.HasPrefix(u, "-") && !strings.HasSuffix(u, "-")
}

// isNumericLabel reports whether a label is a number, one or more
// digits or "0x" followed by hexadecimal digits, as the WHATWG URL
// parser reads the last label of a host.
func isNumericLabel(s string) bool {
	if hex, ok := strings.CutPrefix(s, "0x"); ok {
		return strings.Trim(hex, "0123456789abcdef") == ""
	}
	return s != "" && strings.Trim(s, "0123456789") == ""
}

// lastLabel returns the label after the last dot of a name.
func lastLabel(s string) string {
	return s[strings.LastIndexByte(s, '.')+1:]
}

// joinHostPortUnsafe joins a host and a port without validating either.
// It returns the host alone when the port is empty, and otherwise joins
// them through net.JoinHostPort.
func joinHostPortUnsafe(host, port string) string {
	if port == "" {
		// portless
		return host
	}
	return net.JoinHostPort(host, port)
}

func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
}
