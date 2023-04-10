package utils

// Remove drops an element from a slice and returns a new slice.
func Remove[T comparable](s []T, t T) []T {
	for i, e := range s {
		if e == t {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}
