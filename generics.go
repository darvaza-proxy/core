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
	return doTypeName(reflect.TypeFor[T]())
}

// TypeNameOf returns the name of the type of the value v holds, as %T
// renders it, but with the empty interface spelled any, as [TypeName]
// spells it. A nil v gives <nil>.
func TypeNameOf(v any) string {
	if v == nil {
		return "<nil>"
	}
	return doTypeName(reflect.TypeOf(v))
}

// doTypeName spells t as [TypeName] and [TypeNameOf] do.
func doTypeName(t reflect.Type) string {
	return strings.ReplaceAll(t.String(), "interface {}", "any")
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

// IIf returns one value or the other depending
// on a condition.
//
//revive:disable-next-line:flag-parameter
func IIf[T any](cond bool, yes, no T) T {
	if cond {
		return yes
	}
	return no
}
