package util

import "github.com/mbpolan/openmcs/internal/model"

// region3D represents the dimensions of a single map region.
var region3D = model.Vector3D{
	X: 64,
	Y: 64,
	Z: 1,
}

// mapScale3D represents the scale factor in which map region origins are stored.
var mapScale3D = model.Vector3D{
	X: 8,
	Y: 8,
	Z: 1,
}

// GlobalToRegionLocal scales a position in global coordinates to region local coordinates.
func GlobalToRegionLocal(v model.Vector3D) model.Vector3D {
	return v.Mod(region3D)
}

// GlobalToRegionOrigin scales a position in global coordinates to region origin coordinates.
func GlobalToRegionOrigin(v model.Vector3D) model.Vector3D {
	return model.Vector3D{
		X: ((v.X / region3D.X) * region3D.X) / mapScale3D.X,
		Y: ((v.Y / region3D.Y) * region3D.Y) / mapScale3D.Y,
		Z: v.Z,
	}
}
