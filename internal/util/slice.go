package util

// Remove drops an element from a slice and returns a new slice.
func Remove[T comparable](s []T, t T) []T {
	for i, e := range s {
		if e == t {
			return append(s[:i], s[i+1:]...)
		}
	}

	return s
}

// Contains returns true if a slice contains an element.
func Contains[T comparable](s []T, t T) bool {
	for _, e := range s {
		if e == t {
			return true
		}
	}

	return false
}
