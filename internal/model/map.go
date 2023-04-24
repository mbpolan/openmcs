package model

// Tile is the smallest unit of space on the world map.
type Tile struct {
	Objects   []*WorldObject
	TerrainID int
}

// AddObject places a world object on the tile.
func (t *Tile) AddObject(object *WorldObject) {
	t.Objects = append(t.Objects, object)
}

// Map represents the game world map and its static objects.
type Map struct {
	// Tiles are stored in (z, x, y) coordinates.
	Tiles map[int]map[int]map[int]*Tile
}

// NewMap returns a new world map.
func NewMap() *Map {
	return &Map{
		Tiles: map[int]map[int]map[int]*Tile{},
	}
}

// PutTile initializes a tile at location on the world map.
func (m *Map) PutTile(pos Vector3D) {
	m.Tile(pos)
}

// Tile returns a tile at a location on the world map. If there is no tile, a new one will be initialized.
func (m *Map) Tile(pos Vector3D) *Tile {
	if _, ok := m.Tiles[pos.Z]; !ok {
		m.Tiles[pos.Z] = map[int]map[int]*Tile{}
	}

	if _, ok := m.Tiles[pos.Z][pos.X]; !ok {
		m.Tiles[pos.Z][pos.X] = map[int]*Tile{}
	}

	if _, ok := m.Tiles[pos.Z][pos.X][pos.Y]; !ok {
		m.Tiles[pos.Z][pos.X][pos.Y] = &Tile{}
	}

	return m.Tiles[pos.Z][pos.X][pos.Y]
}

// MapObject is an object that is located on the map.
type MapObject struct {
	ID          int
	Position    Vector3D
	ObjectType  int
	Orientation int
}
