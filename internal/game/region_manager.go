package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
)

type changeEventType int

const (
	changeEventAddGroundItem changeEventType = iota
)

type changeEvent struct {
	Type      changeEventType
	GlobalPos model.Vector3D
}

// RegionManager is responsible for tracking the state of a single, 2D region on the world map.
type RegionManager struct {
	origin          model.Vector3D
	worldMap        *model.Map
	state           []response.Response
	chunkStateIndex map[model.Vector3D]int
	pendingEvents   []*changeEvent
}

// NewRegionManager creates a manager for a 2D region of the map. The z-coordinate will be used to determine which
// plane of the region this manager will be responsible for.
func NewRegionManager(origin model.Vector3D, m *model.Map) *RegionManager {
	return &RegionManager{
		chunkStateIndex: map[model.Vector3D]int{},
		origin:          origin,
		worldMap:        m,
	}
}

// State returns the last computed state of the region, described as a slice of response.Response messages that can
// be sent to a player's client.
func (r *RegionManager) State() []response.Response {
	return r.state
}

// AddGroundItem adds a ground item to a tile.
func (r *RegionManager) AddGroundItem(itemID int, globalPos model.Vector3D) {
	tile := r.worldMap.Tile(globalPos)
	if tile == nil {
		return
	}

	tile.AddItem(itemID)

	r.pendingEvents = append(r.pendingEvents, &changeEvent{
		Type:      changeEventAddGroundItem,
		GlobalPos: globalPos,
	})
}

// Reconcile validates the current state of the region and recomputes its state if a change has occurred. A slice of
// messages will be returned that should be dispatched to players in the region.
func (r *RegionManager) Reconcile() []response.Response {
	// perform an initial state update
	if r.state == nil {
		r.state = r.computeRegion()
		return nil
	}

	var updates []response.Response

	for _, e := range r.pendingEvents {
		switch e.Type {
		case changeEventAddGroundItem:
			// an item was added to a tile
			chunkOrigin := model.Vector3D{
				X: r.origin.X + ((e.GlobalPos.X-r.origin.X)/util.Chunk2D.X)*util.Chunk2D.X,
				Y: r.origin.Y + ((e.GlobalPos.Y-r.origin.Y)/util.Chunk2D.Y)*util.Chunk2D.Y,
				Z: r.origin.Z,
			}

			// recompute the state of the chunk the item was added to
			update := r.computeChunk(chunkOrigin)
			updates = append(updates, update)

			// compute the chunk's new state, and append or update it in the region's state list
			idx, ok := r.chunkStateIndex[chunkOrigin]
			if ok {
				r.state[idx] = update
			} else if update != nil {
				r.chunkStateIndex[chunkOrigin] = len(r.state)
				r.state = append(r.state, update)
			}
		}
	}

	r.pendingEvents = nil
	return updates
}

// computeRegion builds a slice of BatchResponse messages describing the current state of a map region. The origin
// should be the region origin in global coordinates.
func (r *RegionManager) computeRegion() []response.Response {
	var batches []response.Response
	r.chunkStateIndex = map[model.Vector3D]int{}

	// compute batches for each chunk in this region
	for x := 0; x < util.Region3D.X*util.Chunk2D.X; x += util.Chunk2D.X {
		for y := 0; y < util.Region3D.Y*util.Chunk2D.Y; y += util.Chunk2D.Y {
			chunkOrigin := model.Vector3D{
				X: r.origin.X + x,
				Y: r.origin.Y + y,
				Z: r.origin.Z,
			}

			// compute the state of this chunk and mark its index
			chunkBatches := r.computeChunk(chunkOrigin)
			if chunkBatches != nil {
				r.chunkStateIndex[chunkOrigin] = len(batches)
				batches = append(batches, chunkBatches)
			}
		}
	}

	return batches
}

// computeChunk builds a BatchResponse containing the current state of a chunk in a region. The chunkOriginGlobal
// should be the origin of the chunk in global coordinates. If there are no updates for this chunk, then a nil will
// be returned instead.
func (r *RegionManager) computeChunk(chunkOriginGlobal model.Vector3D) response.Response {
	var batched []response.Response

	delete(r.chunkStateIndex, chunkOriginGlobal)
	origin := util.GlobalToRegionLocal(chunkOriginGlobal)

	for x := 0; x <= util.Chunk2D.X; x++ {
		for y := 0; y < util.Chunk2D.Y; y++ {
			// find the tile, if there is one, at this location
			tilePos := model.Vector3D{
				X: chunkOriginGlobal.X + x,
				Y: chunkOriginGlobal.Y + y,
				Z: chunkOriginGlobal.Z,
			}

			// is this tile in bounds?
			if tilePos.X < 0 || tilePos.Y < 0 || tilePos.X > r.worldMap.MaxTile.X || tilePos.Y > r.worldMap.MaxTile.Y {
				continue
			}

			tile := r.worldMap.Tile(tilePos)
			if tile == nil {
				continue
			}

			// find the relative position of this tile with respect to the player's chunk origin. since the x- and y-
			// coordinate offsets can be negative in this loop, we need to ensure that each is translated to be relative
			// to the chunk origin.
			relative := model.Vector2D{
				X: x,
				Y: y,
			}

			// describe ground items at this tile
			for _, item := range tile.ItemIDs {
				batched = append(batched, response.NewShowGroundItemResponse(item, 1, relative))
			}
		}
	}

	if batched == nil {
		return nil
	}

	return response.NewBatchResponse(origin.To2D(), batched)
}
