package model

import (
	"github.com/google/uuid"
	"sync"
)

// TileGroundItem is an instance of an item placed on a tile.
type TileGroundItem struct {
	InstanceUUID uuid.UUID
	ItemID       int
}

// Tile is the smallest unit of space on the world map.
type Tile struct {
	Height     int
	OverlayID  int
	RenderFlag int
	UnderlayID int

	objects     []*WorldObject
	groundItems []*TileGroundItem
	mu          sync.Mutex
}

// AddObject places a world object on the tile.
func (t *Tile) AddObject(object *WorldObject) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.objects = append(t.objects, object)
}

// AddItem adds a ground item to the tile returning its unique instance UUID.
func (t *Tile) AddItem(id int) uuid.UUID {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := &TileGroundItem{
		InstanceUUID: uuid.New(),
		ItemID:       id,
	}

	t.groundItems = append([]*TileGroundItem{item}, t.groundItems...)
	return item.InstanceUUID
}

// GroundItemIDs returns a slice of ground item IDs located on this tile.
func (t *Tile) GroundItemIDs() []int {
	t.mu.Lock()
	defer t.mu.Unlock()

	ids := make([]int, len(t.groundItems))
	for i, item := range t.groundItems {
		ids[i] = item.ItemID
	}

	return ids
}

// RemoveItem removes a ground item that matches the instance UUID. If the item was found and removed, its item ID will
// be returned.
func (t *Tile) RemoveItem(instanceUUID uuid.UUID) *int {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, item := range t.groundItems {
		if item.InstanceUUID == instanceUUID {
			t.groundItems = append(t.groundItems[:i], t.groundItems[i+1:]...)
			return &item.ItemID
		}
	}

	return nil
}

// Clear removes all ground items on the tile.
func (t *Tile) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.groundItems = nil
}

// Map represents the game world map and its static objects.
type Map struct {
	// Tiles are stored in (z, x, y) coordinates.
	Tiles map[int]map[int]map[int]*Tile
	// RegionOrigins enumerate all region origins in global coordinates that are on the map.
	RegionOrigins []Vector3D
	// MaxTile is the position of the tile the furthest on the x- and y-axes.
	MaxTile Vector2D
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

// Tile returns a tile at a location on the world map. If there is no tile, nil will be returned instead.
func (m *Map) Tile(pos Vector3D) *Tile {
	if _, ok := m.Tiles[pos.Z]; !ok {
		return nil
	}

	if _, ok := m.Tiles[pos.Z][pos.X]; !ok {
		return nil
	}

	if _, ok := m.Tiles[pos.Z][pos.X][pos.Y]; !ok {
		return nil
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
