package core

import "strings"

// The runes UTS 46 maps into '%', which separate a zone as '%' does.
const (
	fullWidthPercent = '％'
	smallPercent     = '﹪'
)

// The runes UTS 46 maps into ':', which count as ':' does.
const (
	fullWidthColon = '：'
	smallColon     = '﹕'
	verticalColon  = '︓'
)

// parseHostPort reads the syntax of a host and port, and returns the
// host, its zone and the port, brackets and the zone's separator
// removed, so that the zone stays out of the mapping the host goes
// through. It refuses broken syntax and a zone validZone refuses, and
// leaves what the host and the port hold unchecked. On failure it
// returns the error alone.
//
//revive:disable-next-line:function-result-limit // the host and its zone, beside the port
func parseHostPort(hostPort string) (host, zone, port string, err error) {
	sc := hostPortScanner{src: hostPort}
	switch {
	case sc.done():
		// empty
		err = addrErr(hostPort, "empty address")
	case sc.accept('['):
		// [host]:port [host], read past the '['
		host, zone, port, err = parseHostPortBracketed(&sc)
	case countRunesFunc(hostPort, isColon) > 1:
		// IPv6 without brackets, its zone running to the end
		host, zone, port, err = parseHostPortPlain(&sc, "")
	default:
		// at most one ':', before the port
		host, zone, port, err = parseHostPortPlain(&sc, ":")
	}

	if err == nil && !validZone(zone) {
		// host%a]b [host%a:b] [host%a%b]: a bracket, a ':', a separator
		// or NUL
		host, zone, port = "", "", ""
		err = addrErr(hostPort, "invalid character in zone")
	}

	return host, zone, port, err
}

// parseHostPortBracketed reads what follows the '[': the host, which
// runs to a zone separator or the ']', the zone, which runs to the
// ']', the ']' and a ':' and the port. On failure it returns the
// error alone.
//
//revive:disable-next-line:function-result-limit // the host and its zone, beside the port
func parseHostPortBracketed(sc *hostPortScanner) (host, zone, port string, err error) {
	host = sc.host("]")
	sep, zone := sc.optionalZone("]")
	closed := sc.accept(']')
	port, hasPort := sc.optionalPort()

	switch {
	case !closed:
		// [host
		err = addrErr(sc.src, "missing ']' in address")
	case host == "":
		// []... [%zone]...
		err = addrErr(sc.src, "missing address in brackets")
	case sep != "" && zone == "":
		// [host%]...
		err = addrErr(sc.src, "missing zone after '"+sep+"'")
	case hasPort && port == "":
		// [host]:
		err = addrErr(sc.src, "missing port after ':'")
	case !sc.done():
		// [host]...
		err = addrErr(sc.src, "invalid character after ']'")
	default:
		// [host] [host]:port [host%zone] [host%zone]:port
	}

	if err != nil {
		host, zone, port = "", "", ""
	}
	return host, zone, port, err
}

// parseHostPortPlain reads a host without brackets: the host, which runs
// to a zone separator or a byte in stops, the zone, which runs to a byte
// in stops, and a ':' and the port where stops holds the ':'. On failure
// it returns the error alone.
//
//revive:disable-next-line:function-result-limit // the host and its zone, beside the port
func parseHostPortPlain(sc *hostPortScanner, stops string) (host, zone, port string, err error) {
	host = sc.host(stops)
	sep, zone := sc.optionalZone(stops)
	port, hasPort := sc.optionalPort()

	switch {
	case sep != "" && host == "":
		// %zone
		err = addrErr(sc.src, "missing address before '"+sep+"'")
	case sep != "" && zone == "":
		// host%
		err = addrErr(sc.src, "missing zone after '"+sep+"'")
	case hasPort && port == "":
		// host:
		err = addrErr(sc.src, "missing port after ':'")
	default:
		// host host:port :port host%zone host%zone:port
	}

	if err != nil {
		host, zone, port = "", "", ""
	}
	return host, zone, port, err
}

// countRunesFunc counts the runes of s for which f returns true.
func countRunesFunc(s string, f func(rune) bool) int {
	var n int
	for _, r := range s {
		if f(r) {
			n++
		}
	}
	return n
}

// isColon reports whether a rune is ':' or one UTS 46 maps into one.
func isColon(r rune) bool {
	switch r {
	case ':', fullWidthColon, smallColon, verticalColon:
		return true
	default:
		return false
	}
}

// hostPortScanner reads a host and port left to right: the text and a
// position, kept on the stack, and a method per piece of the grammar,
// which the splits call in its order. It reads bytes, and knows a zone
// separator, '%', '％' or '﹪', by its whole encoding, so no other rune
// is taken for one.
type hostPortScanner struct {
	src string
	pos int
}

// done reports whether the scanner is at the end of the text.
func (sc *hostPortScanner) done() bool {
	return sc.pos >= len(sc.src)
}

// accept advances past the next byte when it is c, and reports whether
// it was.
func (sc *hostPortScanner) accept(c byte) bool {
	if sc.pos < len(sc.src) && sc.src[sc.pos] == c {
		sc.pos++
		return true
	}
	return false
}

// acceptZoneSeparator advances past the next zone separator, when one
// is next, and returns it, or nothing where none was.
func (sc *hostPortScanner) acceptZoneSeparator() string {
	start := sc.pos
	sc.pos += zoneSeparatorLen(sc.src[sc.pos:])
	return sc.src[start:sc.pos]
}

// host reads up to a zone separator, a byte in stops or the end, and
// returns what it read.
func (sc *hostPortScanner) host(stops string) string {
	start := sc.pos
	sc.pos = sc.zoneSeparatorAt(sc.stopAt(stops))
	return sc.src[start:sc.pos]
}

// zone reads up to a byte in stops or the end, and returns what it read.
func (sc *hostPortScanner) zone(stops string) string {
	start := sc.pos
	sc.pos = sc.stopAt(stops)
	return sc.src[start:sc.pos]
}

// optionalZone reads a zone separator and the zone after it, up to a
// byte in stops or the end, and returns both, or nothing where no
// separator was next.
func (sc *hostPortScanner) optionalZone(stops string) (sep, zone string) {
	sep = sc.acceptZoneSeparator()
	if sep == "" {
		return "", ""
	}
	return sep, sc.zone(stops)
}

// optionalPort reads a ':' and the rest of the text as the port, and
// reports whether a ':' was next.
func (sc *hostPortScanner) optionalPort() (port string, found bool) {
	if !sc.accept(':') {
		return "", false
	}
	return sc.rest(), true
}

// rest reads the text left, and returns it.
func (sc *hostPortScanner) rest() string {
	s := sc.src[sc.pos:]
	sc.pos = len(sc.src)
	return s
}

// stopAt returns the position of the first byte in stops from the
// scanner's position on, or the end of the text where none follows.
func (sc *hostPortScanner) stopAt(stops string) int {
	end := len(sc.src)
	for i := 0; i < len(stops); i++ {
		if j := strings.IndexByte(sc.src[sc.pos:end], stops[i]); j >= 0 {
			end = sc.pos + j
		}
	}
	return end
}

// zoneSeparatorAt returns the position of the first zone separator from
// the scanner's position up to end, or end where none comes before it.
// '％' and '﹪' both start with the byte 0xEF, which it looks for before
// the first '%', taking each one found for a separator only where
// zoneSeparatorLen does.
func (sc *hostPortScanner) zoneSeparatorAt(end int) int {
	if j := strings.IndexByte(sc.src[sc.pos:end], '%'); j >= 0 {
		// a '%', which ends the search
		end = sc.pos + j
	}

	for i := sc.pos; i < end; i++ {
		j := strings.IndexByte(sc.src[i:end], 0xEF)
		if j < 0 {
			// no other separator before end
			break
		}
		i += j
		if zoneSeparatorLen(sc.src[i:end]) > 0 {
			// '％' or '﹪'
			return i
		}
	}
	return end
}

// zoneSeparatorLen returns the length of the zone separator s starts
// with, or 0 where it starts with none.
func zoneSeparatorLen(s string) int {
	switch {
	case strings.HasPrefix(s, "%"):
		return 1
	case strings.HasPrefix(s, string(fullWidthPercent)):
		return len(string(fullWidthPercent))
	case strings.HasPrefix(s, string(smallPercent)):
		return len(string(smallPercent))
	default:
		return 0
	}
}
