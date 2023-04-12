package model

// Map represents the game world map and its static objects.
type Map struct {
	Regions map[int]map[int]*MapRegion
}

// MapRegion is an individual chunk of the map.
type MapRegion struct {
	TerrainID int
	Objects   []*MapObject
}

// MapObject is an object that is located on the map.
type MapObject struct {
	ID          int
	Position    Vector3D
	ObjectType  int
	Orientation int
}
