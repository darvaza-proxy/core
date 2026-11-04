package core

import "net/netip"

// NewAddrPortPtrForTest returns a pointer to a copy of ap.
func NewAddrPortPtrForTest(ap netip.AddrPort) *netip.AddrPort {
	return &ap
}
