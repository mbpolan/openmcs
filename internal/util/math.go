package util

func Unit(n int) int {
	return n / Abs(n)
}

func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func Abs(n int) int {
	if n < 0 {
		return n * -1
	}

	return n
}
