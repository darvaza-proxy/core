package core_test

import "syscall"

// connectionReset is the connection reset that *net.OpError counts as
// temporary when it comes from "accept" on this platform.
const connectionReset = syscall.WSAECONNRESET
