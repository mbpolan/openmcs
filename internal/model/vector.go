package model

// Vector2D is a vector in two-dimensional space.
type Vector2D struct {
	X int
	Y int
}

// Vector3D is a vector in three-dimensional space.
type Vector3D struct {
	X int
	Y int
	Z int
}

func (v Vector2D) Sub(w Vector2D) Vector2D {
	return Vector2D{
		X: v.X - w.X,
		Y: v.Y - w.Y,
	}
}

func (v Vector3D) Multiply(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X * w.X,
		Y: v.Y * w.Y,
		Z: v.Z * w.Z,
	}
}

func (v Vector3D) Divide(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X / w.X,
		Y: v.Y / w.Y,
		Z: v.Z / w.Z,
	}
}

func (v Vector3D) Mod(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X % w.X,
		Y: v.Y % w.Y,
		Z: v.Z % w.Z,
	}
}

func (v Vector3D) To2D() Vector2D {
	return Vector2D{
		X: v.X,
		Y: v.Y,
	}
}
