package core

import (
	"reflect"
	"strings"
)

// TypeName returns the name of a type parameter as Go source spells
// it: a concrete type as %T renders a value of it, an interface type
// as itself where %T can only name the dynamic type a value holds,
// and the empty interface as any where %T prints interface {}.
func TypeName[T any]() string {
	s := reflect.TypeFor[T]().String()
	return strings.ReplaceAll(s, "interface {}", "any")
}

// Coalesce returns the first non-zero argument
func Coalesce[T any](opts ...T) T {
	for _, v := range opts {
		if !IsZero(v) {
			return v
		}
	}

	return Zero[T](nil)
}

// revive:disable:flag-parameter

// IIf returns one value or the other depending
// on a condition.
func IIf[T any](cond bool, yes, no T) T {
	// revive:enable:flag-parameter
	if cond {
		return yes
	}
	return no
}
