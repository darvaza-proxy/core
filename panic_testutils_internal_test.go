package core

// PanicPlainForTest panics with a plain value, not a PanicError.
func PanicPlainForTest() error {
	panic(ErrSentinelForTest)
}
