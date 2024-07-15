package core

// Signed is any signed integer type.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is any unsigned integer type.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is any integer type, signed and unsigned.
type Integer interface {
	Signed | Unsigned
}

// Complex is any complex numeric type.
type Complex interface {
	~complex64 | ~complex128
}

// Float is any floating-point type.
type Float interface {
	~float32 | ~float64
}

// Bool is any boolean type.
type Bool interface {
	~bool
}

// String is any string type.
type String interface {
	~string
}

// Ordered is any type that supports order operators.
type Ordered interface {
	Signed | Unsigned | Float | String
}
