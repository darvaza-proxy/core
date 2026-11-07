package core

import "slices"

// AreEqualMaxDepthForTest is the depth AreEqual's walk stops at.
const AreEqualMaxDepthForTest = areEqualMaxDepth

// EqSliceForTest is non-comparable and follows the Equal(T) bool convention.
type EqSliceForTest []int

// Equal reports whether s and o hold the same elements in the same order.
func (s EqSliceForTest) Equal(o EqSliceForTest) bool { return slices.Equal(s, o) }
