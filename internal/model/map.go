package model

import (
	"github.com/google/uuid"
	"sync"
)

// TileGroundItem is an instance of an item placed on a tile.
type TileGroundItem struct {
	InstanceUUID uuid.UUID
	ItemID       int
	Amount       int
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

// AddItem adds a non-stackable ground item to the tile, returning its unique instance UUID.
func (t *Tile) AddItem(id int) uuid.UUID {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := &TileGroundItem{
		InstanceUUID: uuid.New(),
		Amount:       1,
		ItemID:       id,
	}

	t.groundItems = append([]*TileGroundItem{item}, t.groundItems...)
	return item.InstanceUUID
}

// AddStackableItem adds a stackable ground item with a stack amount to the tile, returning its unique instance UUID.
// If a new item was added to the tile, true will be returned in the second tuple element, otherwise false if an
// existing item's stack was updated. If an existing item was updated, the previous stack amount will be returned in
// the third tuple element.
func (t *Tile) AddStackableItem(id, amount int) (uuid.UUID, bool, int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// try to find an existing item we can add the stack amount to
	for _, item := range t.groundItems {
		if item.ItemID == id && int64(item.Amount+amount) < MaxStackableSize {
			// reset the item's instance uuid and add the amount to the stack
			item.InstanceUUID = uuid.New()
			oldAmount := item.Amount
			item.Amount += amount
			return item.InstanceUUID, false, oldAmount
		}
	}

	// otherwise add this as a new item
	item := &TileGroundItem{
		InstanceUUID: uuid.New(),
		Amount:       amount,
		ItemID:       id,
	}

	t.groundItems = append([]*TileGroundItem{item}, t.groundItems...)
	return item.InstanceUUID, true, 0
}

// GroundItems returns a slice of ground items located on this tile.
func (t *Tile) GroundItems() []*TileGroundItem {
	t.mu.Lock()
	defer t.mu.Unlock()

	items := make([]*TileGroundItem, len(t.groundItems))
	for i, item := range t.groundItems {
		items[i] = item
	}

	return items
}

// RemoveItemByID removes the first ground item that matches the item ID. If the item was found and removed, a pointer
// to the TileGroundItem model will be returned. If there are multiple ground items with the same item ID, only the
// first will be removed.
func (t *Tile) RemoveItemByID(id int) *TileGroundItem {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, item := range t.groundItems {
		if item.ItemID == id {
			t.groundItems = append(t.groundItems[:i], t.groundItems[i+1:]...)
			return item
		}
	}

	return nil
}

// RemoveItemByInstanceUUID removes a ground item that matches the instance UUID. If the item was found and removed, its item ID will
// be returned.
func (t *Tile) RemoveItemByInstanceUUID(instanceUUID uuid.UUID) *int {
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
