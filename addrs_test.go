package core

import (
	"testing"
)

func TestParsePort(t *testing.T) {
	tests := []struct {
		input string
		want  uint16
	}{
		{input: "1950", want: uint16(1950)},
		{input: "-89898989", want: uint16(0)},
		{input: "89000", want: uint16(0)},
		{input: "-22", want: uint16(22)},
		{input: "assadssws", want: uint16(0)},
		{input: "0", want: uint16(0)},
	}

	for _, tc := range tests {
		got, _ := ParsePort(tc.input)
		if tc.want != got {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}
