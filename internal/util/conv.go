package util

import "github.com/mbpolan/openmcs/internal/model"

// region3D represents the dimensions of a single map region.
var region3D = model.Vector3D{
	X: 64,
	Y: 64,
	Z: 1,
}

// MapScale3D represents the scale factor in which map region origins are stored.
var MapScale3D = model.Vector2D{
	X: 8,
	Y: 8,
}

// GlobalToRegionLocal scales a position in global coordinates to region local coordinates.
func GlobalToRegionLocal(v model.Vector3D) model.Vector3D {
	return v.Mod(region3D)
}

// GlobalToRegionOrigin scales a position in global coordinates to region origin coordinates.
func GlobalToRegionOrigin(v model.Vector3D) model.Vector3D {
	return model.Vector3D{
		X: ((v.X / region3D.X) * region3D.X) / MapScale3D.X,
		Y: ((v.Y / region3D.Y) * region3D.Y) / MapScale3D.Y,
		Z: v.Z,
	}
}
