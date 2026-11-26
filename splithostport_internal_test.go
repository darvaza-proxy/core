package core

import (
	"net"
	"testing"
)

// Compile-time verification that test case types implement TestCase interface
var _ TestCase = isNumericLabelTestCase{}

type isNumericLabelTestCase struct {
	name     string
	label    string
	expected bool
}

func (tc isNumericLabelTestCase) Name() string {
	return tc.name
}

func (tc isNumericLabelTestCase) Test(t *testing.T) {
	AssertEqual(t, tc.expected, isNumericLabel(tc.label), "numeric")
}

// newIsNumericLabelTestCaseNumber declares a label isNumericLabel reads
// as a number.
func newIsNumericLabelTestCaseNumber(name, label string) isNumericLabelTestCase {
	return isNumericLabelTestCase{
		name:     name,
		label:    MustNotZeroForTest(label),
		expected: true,
	}
}

// newIsNumericLabelTestCaseNotNumber declares a label isNumericLabel
// does not read as a number.
func newIsNumericLabelTestCaseNotNumber(name, label string) isNumericLabelTestCase {
	return isNumericLabelTestCase{
		name:     name,
		label:    label,
		expected: false,
	}
}

func isNumericLabelTestCases() []isNumericLabelTestCase {
	return S(
		// A number is one or more digits, or "0x" followed by
		// hexadecimal digits, none at all included.
		newIsNumericLabelTestCaseNumber("digits", "123"),
		newIsNumericLabelTestCaseNumber("hexadecimal", "0x1f"),
		newIsNumericLabelTestCaseNumber("bare 0x", "0x"),

		// Anything else is not a number, an empty label included.
		newIsNumericLabelTestCaseNotNumber("letters", "name"),
		newIsNumericLabelTestCaseNotNumber("non-hexadecimal 0x", "0xg"),
		newIsNumericLabelTestCaseNotNumber("empty", ""),
	)
}

func TestIsNumericLabel(t *testing.T) {
	RunTestCases(t, isNumericLabelTestCases())
}

// TestAddrErr states what addrErr hands its callers: a *net.AddrError
// carrying its two arguments in the two fields.
func TestAddrErr(t *testing.T) {
	err := addrErr("invalid.address", "test error")

	got := AssertMustErrorAs[*net.AddrError](t, err, "addrErr")
	AssertEqual(t, "invalid.address", got.Addr, "Addr")
	AssertEqual(t, "test error", got.Err, "Err")
}
