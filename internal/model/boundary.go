package model

// Boundary enumerates the cardinal boundaries of a map area, region or chunk.
type Boundary uint8

const (
	BoundaryNone  Boundary = 0
	BoundaryNorth Boundary = 1 << iota
	BoundaryWest
	BoundaryEast
	BoundarySouth
)
