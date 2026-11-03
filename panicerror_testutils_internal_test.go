package core

// NewPanicPayloadForTest builds a *PanicError whose stack starts here,
// apart from any stack Panic would capture.
func NewPanicPayloadForTest() *PanicError {
	return NewPanicError(0, ErrSentinelForTest)
}
