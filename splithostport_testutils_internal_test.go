package core

// MustNotEmptyForTest refuses an empty result in an accepting factory.
// Each subject whose factories call it returns the empty string when it
// rejects its input, so an accepted row declaring one would be a rejected
// row in disguise.
func MustNotEmptyForTest(s string) string {
	if s == "" {
		panic("accepted row declares an empty result")
	}
	return s
}
