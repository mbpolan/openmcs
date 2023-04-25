package model

// Tile is the smallest unit of space on the world map.
type Tile struct {
	Height     int
	Objects    []*WorldObject
	ItemIDs    []int
	OverlayID  int
	RenderFlag int
	UnderlayID int
}

// AddObject places a world object on the tile.
func (t *Tile) AddObject(object *WorldObject) {
	t.Objects = append(t.Objects, object)
}

// AddItem adds a ground item to the tile.
func (t *Tile) AddItem(id int) {
	t.ItemIDs = append([]int{id}, t.ItemIDs...)
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

// SetTile puts a tile at a location. Any existing tile will be replaced.
func (m *Map) SetTile(pos Vector3D, tile *Tile) {
	m.ensurePathToTile(pos)

	m.Tiles[pos.Z][pos.X][pos.Y] = tile
}

// Tile returns a tile at a location on the world map. If there is no tile, a new one will be initialized.
func (m *Map) Tile(pos Vector3D) *Tile {
	m.ensurePathToTile(pos)

	if _, ok := m.Tiles[pos.Z][pos.X][pos.Y]; !ok {
		m.Tiles[pos.Z][pos.X][pos.Y] = &Tile{}
	}

	return m.Tiles[pos.Z][pos.X][pos.Y]
}

// ensurePathToTile initializes the tile map to a particular tile.
func (m *Map) ensurePathToTile(pos Vector3D) {
	if _, ok := m.Tiles[pos.Z]; !ok {
		m.Tiles[pos.Z] = map[int]map[int]*Tile{}
	}

	if _, ok := m.Tiles[pos.Z][pos.X]; !ok {
		m.Tiles[pos.Z][pos.X] = map[int]*Tile{}
	}
}
