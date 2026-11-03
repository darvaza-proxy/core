package core

import "errors"

// testHello is a string value the tests share as a payload and an
// expected value.
const testHello = "hello"

var (
	// ErrSentinelForTest is the known error the tests pass in and
	// expect back, matched by identity.
	ErrSentinelForTest = errors.New("sentinel error")

	// ErrUnrelatedForTest is never put into any chain under test. The
	// errors.Is assertions pair a match against it as a non-match, so
	// a chain that answered true to everything would fail.
	ErrUnrelatedForTest = errors.New("unrelated error")
)
