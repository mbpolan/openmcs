package util

import "github.com/mbpolan/openmcs/internal/model"

// Region3D represents the dimensions of a single map region.
var Region3D = model.Vector3D{
	X: 64,
	Y: 64,
	Z: 1,
}

// Chunk2D represents the dimensions of a single region chunk.
var Chunk2D = model.Vector2D{
	X: 8,
	Y: 8,
}

// RegionBoundary2D represents the amount of tiles that comprise the overlapping boundary between regions.
var RegionBoundary2D = model.Vector2D{
	X: 6,
	Y: 6,
}

// GlobalToRegionLocal scales a position in global coordinates to region local coordinates.
func GlobalToRegionLocal(v model.Vector3D) model.Vector3D {
	return v.Mod(Region3D)
}

// GlobalToRegionOrigin scales a position in global coordinates to region origin coordinates.
func GlobalToRegionOrigin(v model.Vector3D) model.Vector3D {
	return model.Vector3D{
		X: ((v.X / Region3D.X) * Region3D.X) / Chunk2D.X,
		Y: ((v.Y / Region3D.Y) * Region3D.Y) / Chunk2D.Y,
		Z: v.Z,
	}
}

// GlobalToRegionGlobal translates a position in global coordinates to region origin global coordinates.
func GlobalToRegionGlobal(v model.Vector3D) model.Vector3D {
	return model.Vector3D{
		X: (v.X / Region3D.X) * Region3D.X,
		Y: (v.Y / Region3D.Y) * Region3D.Y,
		Z: v.Z,
	}
}
