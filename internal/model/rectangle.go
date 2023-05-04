package model

// Rectangle is a rectangle aligned on the x- and y-axes.
type Rectangle struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

// MakeRectangle returns a Rectangle initialized with a top-left and bottom-right boundary point.
func MakeRectangle(x1, y1, x2, y2 int) Rectangle {
	return Rectangle{
		X1: x1,
		Y1: y1,
		X2: x2,
		Y2: y2,
	}
}
