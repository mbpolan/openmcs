package util

import "github.com/mbpolan/openmcs/internal/model"

// Region3D represents the dimensions of a single map region.
var Region3D = model.Vector3D{
	X: 64,
	Y: 64,
	Z: 1,
}

// Area2D represents the dimensions of a single map area.
var Area2D = model.Vector2D{
	X: 13,
	Y: 13,
}

// Chunk2D represents the dimensions of a single region chunk.
var Chunk2D = model.Vector2D{
	X: 8,
	Y: 8,
}

// RegionBoundary2D represents the amount of tiles that comprise the overlapping boundary between regions.
var RegionBoundary2D = model.Vector2D{
	X: 16,
	Y: 16,
}

// ClientChunkArea2D represents the amount of chunks that comprise the area drawn by the client.
var ClientChunkArea2D = model.Vector2D{
	X: 13,
	Y: 13,
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

// RegionOriginToGlobal translates a region origin, in region origin coordinates, to global coordinates.
func RegionOriginToGlobal(v model.Vector2D) model.Vector3D {
	return model.Vector3D{
		X: v.X * Chunk2D.X,
		Y: v.Y * Chunk2D.Y,
		Z: 0,
	}
}

// RegionGlobalToClientBase translates a region origin, in global coordinates, to the base coordinates used by the
// game client.
func RegionGlobalToClientBase(v model.Vector3D) model.Vector3D {
	return model.Vector3D{
		X: (((v.X / Region3D.X) * Chunk2D.X) - 6) * Chunk2D.X,
		Y: (((v.Y / Region3D.Y) * Chunk2D.Y) - 6) * Chunk2D.Y,
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
