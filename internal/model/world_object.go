package model

// WorldObject is an entity that can appear on the game world.
type WorldObject struct {
	ID              int
	Description     string
	Name            string
	Walkable        bool
	Actions         []string
	UnwalkableSolid bool
	Translation     Vector3D
	Size            Vector2D
	Scale           Vector3D
	FaceID          int
	MapSceneID      int
	Shadowless      bool
	Rotated         bool
	OffsetAmplified bool
	AdjustToTerrain bool
	DelayedShading  bool
	HasActions      bool
	Solid           bool
	Wall            bool
	Static          bool
	VariableID      int
	ConfigID        int
	ChildIDs        []int
}
