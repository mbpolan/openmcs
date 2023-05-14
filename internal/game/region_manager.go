package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
	"math"
	"sync"
)

// chunkState is a container for the state of a particular chunk.
type chunkState struct {
	// boundary defines which boundary of the region this chunk lies on, if any.
	boundary model.Boundary
	// state is the current state of the chunk.
	state response.Response
	// relative is the offset of this chunk relative to the region origin.
	relative model.Vector2D
}

// RegionManager is responsible for tracking the state of a single, 2D region on the world map. A region is defined by
// a square of size util.Region3D centered about an origin, plus additional tiles on each boundary equal to
// util.Chunk2D * 2. Therefore, the entire span of tiles for a RegionManager is util.Area2D.
type RegionManager struct {
	// initialized flags if the RegionManager has performed an initial state update.
	initialized bool
	// origin is the position, in global coordinates, at the center of the region.
	origin model.Vector3D
	// clientBaseRegion is the top-left position, in global coordinates, of the region plus area padding relative to
	// the client's base coordinates.
	clientBaseArea model.Rectangle
	// clientBaseRegion is the top-left position, in global coordinates, of the region relative to the client's
	// base coordinates.
	clientBaseRegion model.Rectangle
	// worldMap is a pointer to the parent model.Map model.
	worldMap *model.Map
	// state is a memoized slice of each chunk's state.
	state []*chunkState
	// chunkRelative is a map of chunk origins, in global coordinates, to their offsets from the region origin.
	chunkRelative map[model.Vector3D]model.Vector2D
	// chunkStates is a map of chunk origins, in global coordinates, to their last computed state.
	chunkStates map[model.Vector3D]*chunkState
	// pendingEvents is a slice of deltas that have occurred to this region's state that need to be reconciled.
	pendingEvents []*changeDelta
	// mu is a mutex for volatile struct fields.
	mu sync.Mutex
}

// NewRegionManager creates a manager for a 2D region of the map centered around an origin in global coordinates.
// The z-coordinate will be used to determine which plane of the region this manager will be responsible for.
func NewRegionManager(origin model.Vector3D, m *model.Map) *RegionManager {
	// compute the top-left coordinates of the region, including the area padding relative to client base coordinates
	clientBaseArea := model.Rectangle{
		X1: origin.X - (util.Area2D.X/2)*util.Chunk2D.X,
		Y1: origin.Y + ((util.Area2D.Y/2)+1)*util.Chunk2D.Y,
		X2: origin.X + ((util.Area2D.X/2)+1)*util.Chunk2D.X,
		Y2: origin.Y - (util.Area2D.Y/2)*util.Chunk2D.Y,
	}

	// compute the top-left coordinates of the region, not including area padding relative to client base coordinates
	clientBaseRegion := model.Rectangle{
		X1: origin.X - (util.Chunk2D.X/2)*util.Chunk2D.X,
		Y1: origin.Y + ((util.Chunk2D.Y/2)+1)*util.Chunk2D.Y,
		X2: origin.X + ((util.Chunk2D.X/2)+1)*util.Chunk2D.X,
		Y2: origin.Y - (util.Chunk2D.Y/2)*util.Chunk2D.Y,
	}

	mgr := &RegionManager{
		chunkRelative:    map[model.Vector3D]model.Vector2D{},
		chunkStates:      map[model.Vector3D]*chunkState{},
		origin:           origin,
		clientBaseArea:   clientBaseArea,
		clientBaseRegion: clientBaseRegion,
		worldMap:         m,
	}

	return mgr
}

// Contains returns true if a position, in global coordinates, is managed by this manager.
func (r *RegionManager) Contains(globalPos model.Vector3D) bool {
	return globalPos.X >= r.clientBaseArea.X1 && globalPos.X <= r.clientBaseArea.X2 &&
		globalPos.Y >= r.clientBaseArea.Y2 && globalPos.Y <= r.clientBaseArea.Y1
}

// State returns the last computed state of the region, described as a slice of response.Response messages that can
// be sent to a player's client.
func (r *RegionManager) State(trim model.Boundary) []response.Response {
	var state []response.Response
	for _, st := range r.state {
		if st.state == nil || st.boundary&trim != 0 {
			continue
		}

		state = append(state, st.state)
	}

	return state
}

// MarkGroundItemAdded informs the region manager that a ground Item with a stack amount was placed on a tile.
func (r *RegionManager) MarkGroundItemAdded(itemID, amount int, globalPos model.Vector3D) {
	// track this change to the region state
	r.addDelta(&changeDelta{
		eventType: changeEventAddGroundItem,
		globalPos: globalPos,
		items: []changeDeltaItem{
			{
				itemID: itemID,
				amount: amount,
			},
		},
	})
}

// MarkGroundItemUpdated informs the region manager that a ground Item's stack amount was updated.
func (r *RegionManager) MarkGroundItemUpdated(itemID, oldAmount, newAmount int, globalPos model.Vector3D) {
	// track this change to the region state
	r.addDelta(&changeDelta{
		eventType: changeEventUpdateGroundItem,
		globalPos: globalPos,
		items: []changeDeltaItem{
			{
				itemID:    itemID,
				amount:    newAmount,
				oldAmount: oldAmount,
			},
		},
	})
}

// MarkGroundItemsCleared informs the region manager that all ground items on a tile have been removed.
func (r *RegionManager) MarkGroundItemsCleared(itemIDs []int, globalPos model.Vector3D) {
	items := make([]changeDeltaItem, len(itemIDs))
	for i, id := range itemIDs {
		items[i] = changeDeltaItem{
			itemID: id,
		}
	}

	r.addDelta(&changeDelta{
		eventType: changeEventRemoveGroundItem,
		globalPos: globalPos,
		items:     items,
	})
}

// Reconcile validates the current state of the region and recomputes its state if a change has occurred. A slice of
// messages will be returned that should be dispatched to players in the region.
func (r *RegionManager) Reconcile() []response.Response {
	// perform an initial state update
	if !r.initialized {
		r.initialized = true
		r.refreshRegion()
		return nil
	}

	r.mu.Lock()

	// track updates by chunk origin, relative to the region origin
	updates := map[model.Vector2D][]response.Response{}

	for _, e := range r.pendingEvents {
		// find the chunk where this change occurred
		chunkOrigin, tileRelative := r.globalToChunkOriginAndRelative(e.globalPos)
		chunkRelative := r.chunkRelative[chunkOrigin]

		switch e.eventType {
		case changeEventAddGroundItem:
			// one or more ground items were added to a tile
			for _, item := range e.items {
				updates[chunkRelative] = append(updates[chunkRelative], response.NewShowGroundItemResponse(item.itemID, item.amount, tileRelative))
			}

		case changeEventRemoveGroundItem:
			// one or more ground items on a tile were removed
			for _, item := range e.items {
				updates[chunkRelative] = append(updates[chunkRelative], response.NewRemoveGroundItemResponse(item.itemID, tileRelative))
			}

		case changeEventUpdateGroundItem:
			// one or more ground Item stack amounts has changed
			for _, item := range e.items {
				updates[chunkRelative] = append(updates[chunkRelative], response.NewUpdateGroundItemResponse(item.itemID, item.oldAmount, item.amount, tileRelative))
			}

		default:
			continue
		}

		// recompute the state of the tile where the change occurred
		// TODO: can this be optimized to only update the tile itself?
		newState := r.computeChunk(chunkOrigin, chunkRelative)
		if newState == nil {
			delete(r.chunkStates, chunkOrigin)
		} else {
			r.chunkStates[chunkOrigin] = newState
		}

		// since this chunk's state might have changed, we need to synchronize the region's memoized state
		r.syncOverallState()
	}

	// clear out all pending events
	r.pendingEvents = nil
	r.mu.Unlock()

	// convert each chunk's updates into a single batched update per chunk
	var batchedUpdates []response.Response
	for chunk, chunkUpdates := range updates {
		batchedUpdates = append(batchedUpdates, response.NewBatchResponse(chunk, chunkUpdates))
	}

	return batchedUpdates
}

// boundaryForChunk returns a bitmask describing if a chunk origin, in global coordinates, falls into one or more of
// the area padding boundaries around a region.
func (r *RegionManager) boundaryForChunk(origin model.Vector3D) model.Boundary {
	// determine if this chunk falls into the area padding around the region. we also need to add another chunk's
	// worth of padding since the client renders extra tiles further ahead of said boundary
	boundary := model.BoundaryNone
	if origin.X <= r.clientBaseRegion.X1-util.Chunk2D.X {
		boundary |= model.BoundaryWest
	} else if origin.X >= r.clientBaseRegion.X2+util.Chunk2D.X {
		boundary |= model.BoundaryEast
	}

	if origin.Y <= r.clientBaseRegion.Y2+util.Chunk2D.X {
		boundary |= model.BoundarySouth
	} else if origin.Y >= r.clientBaseRegion.Y1-util.Chunk2D.X {
		boundary |= model.BoundaryNorth
	}

	return boundary
}

// globalToChunkOriginAndRelative translates a position in global coordinates to the origin of the containing chunk,
// in global coordinates and a relative offset to that position from said origin.
func (r *RegionManager) globalToChunkOriginAndRelative(globalPos model.Vector3D) (model.Vector3D, model.Vector2D) {
	dx := math.Floor(float64(globalPos.X-r.origin.X) / float64(util.Chunk2D.X))
	dy := math.Floor(float64(globalPos.Y-r.origin.Y) / float64(util.Chunk2D.Y))

	chunkOrigin := model.Vector3D{
		X: r.origin.X + int(dx)*util.Chunk2D.X,
		Y: r.origin.Y + int(dy)*util.Chunk2D.Y,
		Z: r.origin.Z,
	}

	// compute the relative offsets to the tile with respect to the chunk origin
	relative := model.Vector2D{
		X: util.Abs(globalPos.X - chunkOrigin.X),
		Y: util.Abs(globalPos.Y - chunkOrigin.Y),
	}

	return chunkOrigin, relative
}

// addDelta adds a change that occurred in the region that should be reported the next time the manager reconciles
// its changes.
func (r *RegionManager) addDelta(delta *changeDelta) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pendingEvents = append(r.pendingEvents, delta)
}

// syncOverallState refreshes the memoized state so that it matches the state of each chunk. You should call this
// method if any of the chunk states have changed, and have not yet been synced with the region's state.
func (r *RegionManager) syncOverallState() {
	r.state = nil
	for _, chunkState := range r.chunkStates {
		r.state = append(r.state, chunkState)
	}
}

// refreshRegion refreshes the current state of this map region, including all the individual chunks that comprise it.
func (r *RegionManager) refreshRegion() {
	r.state = nil
	r.chunkStates = map[model.Vector3D]*chunkState{}

	// compute batches for each chunk in this region
	for x := -util.Area2D.X / 2; x <= util.Area2D.X/2; x++ {
		for y := -util.Area2D.Y / 2; y <= util.Area2D.Y/2; y++ {
			// compute the top-left origin of this chunk in global coordinates
			chunkOrigin := model.Vector3D{
				X: r.origin.X + (x * util.Chunk2D.X),
				Y: r.origin.Y + (y * util.Chunk2D.Y),
				Z: r.origin.Z,
			}

			if chunkOrigin.X < 0 || chunkOrigin.Y < 0 {
				continue
			}

			// compute the relative location of this chunk with respect to the client's base coordinates
			// since we're iterating +/- relative to the origin, we need to translate each x,y offset so that it's
			// instead relative to the top-left coordinate
			chunkRelative := model.Vector2D{
				X: (util.Chunk2D.X * 2) + (x+util.Chunk2D.X/2)*util.Chunk2D.X,
				Y: (util.Chunk2D.Y * 2) + (y+util.Chunk2D.Y/2)*util.Chunk2D.Y,
			}

			r.chunkRelative[chunkOrigin] = chunkRelative

			// compute the state of this chunk and add it to the overall region state
			state := r.computeChunk(chunkOrigin, chunkRelative)
			if state == nil {
				delete(r.chunkStates, chunkOrigin)
			} else {
				r.chunkStates[chunkOrigin] = state
				r.state = append(r.state, state)
			}
		}
	}
}

// computeChunk builds a response.Response slice containing the current state of a chunk in a region. The
// chunkOriginGlobal should be the origin of the chunk in global coordinates. If there are no updates for this chunk,
// then a nil will be returned instead.
func (r *RegionManager) computeChunk(chunkOriginGlobal model.Vector3D, chunkRelative model.Vector2D) *chunkState {
	var batched []response.Response

	for x := 0; x < util.Chunk2D.X; x++ {
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

			tileRelative := model.Vector2D{
				X: x,
				Y: y,
			}

			// compute the state of the tile
			tileState := r.computeTile(tilePos, tileRelative)
			if len(tileState) > 0 {
				batched = append(batched, tileState...)
			}
		}
	}

	// only compute the state if there is at least one tile with state
	if len(batched) == 0 {
		return nil
	}

	boundary := r.boundaryForChunk(chunkOriginGlobal)
	state := response.NewBatchResponse(chunkRelative, batched)

	return &chunkState{
		boundary: boundary,
		relative: chunkRelative,
		state:    state,
	}
}

// computeTile builds a response.Response slice containing the current state of a tile. The tilePosGlobal should be
// the coordinates of the tile in global coordinates. The relative coordinates are the x- and y-coordinate offsets to
// this tile, relative to the origin. If a tile does not exist at the specified location, or if there are no state
// updates for this tile, nil will be returned.
func (r *RegionManager) computeTile(tilePosGlobal model.Vector3D, relative model.Vector2D) []response.Response {
	tile := r.worldMap.Tile(tilePosGlobal)
	if tile == nil {
		return nil
	}

	// describe ground items at this tile
	var batched []response.Response
	for _, item := range tile.GroundItems() {
		batched = append(batched, response.NewShowGroundItemResponse(item.ItemID, item.Amount, relative))
	}

	return batched
}
