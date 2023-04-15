package util

func Unit(n int) int {
	return n / Abs(n)
}

func Abs(n int) int {
	if n < 0 {
		return n * -1
	}

	return n
}
