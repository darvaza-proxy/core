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
	// a failed parse returns the error alone
	host, zone, port, err := parseHostPort(hostPort)
	if err != nil {
		// bad syntax
		return "", err
	}

	port, ok := makePort(port, defaultPort)
	if !ok {
		// bad port, or 0 in any spelling
		return "", addrErr(hostPort, "invalid port")
	}

	if host == "" {
		// port alone
		return formatValidHostPort(netip.Addr{}, "", port), nil
	}

	addr, _, name, ok := mapValidHostForms(host, zone)
	switch {
	case !ok:
		// bad host or zone
		return "", addrErr(hostPort, "invalid address")
	case addr.IsValid():
		// IP address and its zone, as given in host where it has none
		return formatValidHostPort(addr.WithZone(zone), host, port), nil
	default:
		// name, in Unicode
		return formatValidHostPort(addr, name, port), nil
	}
}

// makePort returns the port MakeHostPort joins, in canonical decimal
// form: port, or defaultPort where port is empty, or nothing where both
// are unset. A port that does not parse, or parses to 0, comes back as
// nothing but false.
func makePort(port string, defaultPort uint16) (string, bool) {
	switch {
	case port != "":
		s, ok := canonicalPort(port)
		if !ok || s == "0" {
			// bad port, or 0 in any spelling
			return "", false
		}
		return s, true
	case defaultPort != 0:
		// default port
		return formatPort(defaultPort), true
	default:
		// neither port nor default, portless
		return "", true
	}
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
		return formatHostPort(netip.Addr{}, "", port)
	}

	// the host whole, a zone in it read onto the address by ParseAddr
	addr, _, name, ok := mapValidHostForms(host, "")
	switch {
	case !ok:
		// bad host
		return "", addrErr(host, "invalid host")
	case addr.IsValid():
		// IP address, as given in host
		return formatHostPort(addr, host, port)
	default:
		// name, in Unicode
		return formatHostPort(addr, name, port)
	}
}

// formatHostPort checks a port, and joins it to a valid host as
// formatValidHostPort does, the port in canonical decimal form. It
// refuses a port that does not parse, and the error names the host, as
// formatValidHostPort writes it, and the port as passed.
func formatHostPort(addr netip.Addr, host, port string) (string, error) {
	if port == "" {
		// portless
		return formatValidHostPort(addr, host, ""), nil
	}

	s, ok := canonicalPort(port)
	if !ok {
		// bad port
		return "", addrErr(formatValidHostPort(addr, host, port), "invalid port")
	}

	return formatValidHostPort(addr, host, s), nil
}

// formatValidHostPort joins a valid host and a port, either of them
// possibly empty, in one go, and returns the host alone when the port is
// empty. An IP address comes as addr, and is written in its canonical
// text, in brackets where it is IPv6 and a port follows; where host
// already holds that text, host stands for it. Anything else, a name or
// nothing, comes as host, and is joined as it is.
func formatValidHostPort(addr netip.Addr, host, port string) string {
	var open, colon string
	switch {
	case port == "":
		// portless
	case addr.Is6():
		// IPv6 address before a port
		open, colon = "[", "]:"
	default:
		// IPv4 address, name or nothing before a port
		colon = ":"
	}

	if !addr.IsValid() {
		// a name, or nothing
		return host + colon + port
	}

	// the longest text of an IP address without a zone, which a zone
	// may outgrow
	var buf [len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")]byte
	text := addr.AppendTo(buf[:0])
	if string(text) == host {
		// as written
		return open + host + colon + port
	}

	// address written again
	return open + string(text) + colon + port
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
//   - Returns a valid zone of an IPv6 address unmodified, after a '%',
//     or a '％' or '﹪', which UTS 46 maps into one; the address before
//     it goes through the mapping, the zone does not. A zone holding a
//     bracket, a ':', a separator or NUL is refused
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
//   - SplitHostPort("[fe80::1%ETH0]:80") → ("fe80::1%ETH0", "80", nil)
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

	// a failed parse returns neither host, zone nor port
	host, zone, port, err := parseHostPort(hostPort)
	if port != "" {
		// canonicalPort returns no port for a bad one
		port, ok = canonicalPort(port)
		if !ok {
			// bad port, leaving no host to check
			host, err = "", addrErr(hostPort, "invalid port")
		}
	}

	if host != "" {
		// validHostForms returns nothing for a bad host or zone
		addr, ascii, unicode, ok = validHostForms(host, zone)
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
//   - Returns a valid zone of an IPv6 address unmodified, as
//     SplitHostPort does
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

	// a failed parse returns neither host, zone nor port
	host, zone, sPort, err := parseHostPort(addrPort)
	if sPort != "" {
		// parsePort returns 0 for a bad port
		port, err = parsePort(sPort)
		if err != nil {
			// bad port, leaving no address to check
			host, err = "", addrErr(addrPort, "invalid port")
		}
	}

	if host != "" {
		// validAddr returns the zero netip.Addr for a bad address or zone
		addr, ok = validAddr(host, zone)
		if !ok {
			// bad address
			port, err = 0, addrErr(addrPort, "invalid address")
		}
	}

	return addr, port, err
}

// validZone reports whether a zone holds no bracket, ':', zone
// separator or NUL. A bracket leaves an address joined in brackets that
// does not split back, no Linux interface name holds a ':', RFC 4007
// keeps the separator out of a zone, and RFC 9844 keeps NUL out.
func validZone(zone string) bool {
	return !strings.ContainsFunc(zone, invalidZoneRune)
}

// invalidZoneRune reports whether a rune is one a zone may not hold: a
// bracket, a ':', NUL, or a zone separator, '%', '％' or '﹪'.
func invalidZoneRune(r rune) bool {
	switch r {
	case '[', ']', ':', 0, '%', fullWidthPercent, smallPercent:
		return true
	default:
		return false
	}
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
// decimal form: as given, or written again where it starts with a '0'
// beyond port 0 itself, so a spelling such as "0080" comes back as "80".
func canonicalPort(s string) (string, bool) {
	n, err := parsePort(s)
	switch {
	case err != nil:
		// bad port
		return "", false
	case len(s) > 1 && s[0] == '0':
		// leading zeros
		return formatPort(n), true
	default:
		// canonical as given
		return s, true
	}
}

// validHostForms validates a host and its zone, as mapValidHostForms
// does, and returns the host in its forms. An IP address comes back
// with the zone on it, and its canonical text, zone included, as both
// ascii and unicode. A name comes back as mapValidHostForms returns it.
//
//revive:disable-next-line:function-result-limit // the host in its forms
func validHostForms(s, zone string) (addr netip.Addr, ascii, unicode string, ok bool) {
	addr, ascii, unicode, ok = mapValidHostForms(s, zone)
	if addr.IsValid() {
		// IP address and its zone, written once the zone is on
		addr = addr.WithZone(zone)
		ascii = addr.String()
		unicode = ascii
	}
	return addr, ascii, unicode, ok
}

// mapValidHostForms maps a host as mapHost does, and validates it as an
// IP address or a name, and a zone after it, which only an IPv6 address
// may have. Where the host is an IP address, or UTS 46 maps it into
// one, it returns the address, the zone not put on it, and no text. A name
// comes back with the zero netip.Addr, in the forms validNameForms
// returns. A host it refuses, or a zone after an IPv4 address or a
// name, comes back as nothing but false.
//
//revive:disable-next-line:function-result-limit // the host in its forms
func mapValidHostForms(s, zone string) (addr netip.Addr, ascii, unicode string, ok bool) {
	addr, ascii, ok = mapHost(s)
	switch {
	case !ok:
		// refused by idna, nothing to return
	case addr.IsValid() && !zoneMayFollow(addr, zone):
		// a zone after an IPv4 address
		addr, ok = netip.Addr{}, false
	case addr.IsValid():
		// IP address, without the zone, no text
	case zone != "":
		// a zone after a name
		ascii, ok = "", false
	default:
		// name
		ascii, unicode, ok = validNameForms(ascii)
	}

	return addr, ascii, unicode, ok
}

// validNameForms validates a name, in the ASCII form UTS 46 maps it
// into, as a host name, and returns it in that form and decoded into
// Unicode, which only a name with a label in punycode needs. A name it
// refuses comes back as nothing but false.
func validNameForms(name string) (ascii, unicode string, ok bool) {
	switch {
	case !validHostName(name):
		// not a host name
		return "", "", false
	case !hasALabel(name):
		// name in plain ASCII, its own Unicode form
		return name, name, true
	default:
		// international name
		var err error
		unicode, err = nameProfile.ToUnicode(name)
		if err != nil {
			// idna refuses, once decoded, some names ToASCII
			// accepts, the bidi rule among others
			return "", "", false
		}
		return name, unicode, true
	}
}

// validAddr parses a host as an IP address, or as the IP address
// UTS 46 maps it into, and returns it with the zone on it, as given. A
// host that is no IP address, or a zone that may not follow it, as
// zoneMayFollow reports, comes back as the zero netip.Addr and false.
func validAddr(s, zone string) (netip.Addr, bool) {
	addr, _, _ := mapHost(s)
	if !addr.IsValid() || !zoneMayFollow(addr, zone) {
		// a name, refused by idna, or a zone after an IPv4 address
		return netip.Addr{}, false
	}
	return addr.WithZone(zone), true
}

// zoneMayFollow reports whether a zone may follow an IP address: none,
// or one after an IPv6 address.
func zoneMayFollow(addr netip.Addr, zone string) bool {
	return zone == "" || addr.Is6()
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

func addrErr(addr, why string) error {
	return &net.AddrError{Err: why, Addr: addr}
}
