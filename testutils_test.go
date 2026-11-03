package core_test

// testHello is a string value the tests share as a payload and an
// expected value.
const testHello = "hello"

// namedString is a string by kind only, which NewPanicError does not
// convert to an error.
type namedString string

// mockStringer is a fmt.Stringer whose String returns its value.
type mockStringer struct {
	value string
}

func newMockStringer(value string) mockStringer {
	return mockStringer{value: value}
}

func (ms mockStringer) String() string {
	return ms.value
}
