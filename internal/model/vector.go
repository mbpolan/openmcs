package model

import "fmt"

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

// Sub subtracts another vector from this one, returning a new Vector2D.
func (v Vector2D) Sub(w Vector2D) Vector2D {
	return Vector2D{
		X: v.X - w.X,
		Y: v.Y - w.Y,
	}
}

func (v Vector2D) String() string {
	return fmt.Sprintf("(%d,%d)", v.X, v.Y)
}

// Add adds another vector to this one, returning a new Vector3D.
func (v Vector3D) Add(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X + w.X,
		Y: v.Y + w.Y,
		Z: v.Z + w.Z,
	}
}

// Sub subtracts another vector from this one, returning a new Vector3D.
func (v Vector3D) Sub(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X - w.X,
		Y: v.Y - w.Y,
		Z: v.Z - w.Z,
	}
}

// Multiply multiplies another vector by this one, returning a new Vector3D.
func (v Vector3D) Multiply(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X * w.X,
		Y: v.Y * w.Y,
		Z: v.Z * w.Z,
	}
}

// Divide divides this vector by another one, returning a new Vector3D.
func (v Vector3D) Divide(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X / w.X,
		Y: v.Y / w.Y,
		Z: v.Z / w.Z,
	}
}

// Mod computes the modulo of this vector and another, returning a new Vector3D.
func (v Vector3D) Mod(w Vector3D) Vector3D {
	return Vector3D{
		X: v.X % w.X,
		Y: v.Y % w.Y,
		Z: v.Z % w.Z,
	}
}

// To2D returns a Vector2D with only the two-dimensional components of this vector.
func (v Vector3D) To2D() Vector2D {
	return Vector2D{
		X: v.X,
		Y: v.Y,
	}
}

func (v Vector3D) String() string {
	return fmt.Sprintf("(%d,%d,%d)", v.X, v.Y, v.Z)
}
