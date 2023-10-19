package util

func Unit(n int) int {
	return n / Abs(n)
}

func Clamp(n float32, low, high int) float32 {
	if int(n) < low {
		return float32(low)
	} else if int(n) > high {
		return float32(high)
	}

	return n
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
