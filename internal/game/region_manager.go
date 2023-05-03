package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
	"sync"
	"time"
)

// changeEventType enumerates possible mutation events to a tile.
type changeEventType int

const (
	changeEventAddGroundItem changeEventType = iota
	changeEventRemoveGroundItem
)

// changeDelta is a mutation to a tile that should be tracked.
type changeDelta struct {
	eventType changeEventType
	globalPos model.Vector3D
	itemIDs   []int
}

// chunkState is a container for the state of a particular chunk.
type chunkState struct {
	boundary model.Boundary
	state    response.Response
}

// RegionManager is responsible for tracking the state of a single, 2D region on the world map. A region is defined by
// a square of size util.Region3D centered about an origin, plus additional tiles on each boundary equal to
// util.Chunk2D * 2. Therefore, the entire span of tiles for a RegionManager is util.Area2D.
type RegionManager struct {
	changeChan    chan model.Vector3D
	doneChan      chan bool
	resetChan     chan bool
	origin        model.Vector3D
	topLeft       model.Vector2D
	bottomRight   model.Vector2D
	worldMap      *model.Map
	state         []*chunkState
	chunkStates   map[model.Vector3D]*chunkState
	pendingEvents []*changeDelta
	scheduler     *Scheduler
	mu            sync.Mutex
}

// NewRegionManager creates a manager for a 2D region of the map centered around an origin in global coordinates.
// The z-coordinate will be used to determine which plane of the region this manager will be responsible for. A
// changeChan will be written to when an internal event causes the state of the region to change, containing the origin
// of the region this manager is overseeing.
func NewRegionManager(origin model.Vector3D, m *model.Map, changeChan chan model.Vector3D) *RegionManager {
	// compute the top-left coordinates of the region, not including area padding
	topLeft := model.Vector2D{
		X: origin.X - (util.Chunk2D.X/2)*util.Chunk2D.X,
		Y: origin.Y - (util.Chunk2D.Y/2)*util.Chunk2D.Y,
	}

	// compute the bottom right coordinates of the region, not including area padding
	bottomRight := model.Vector2D{
		X: origin.X + ((util.Chunk2D.X / 2) * util.Chunk2D.X) - 1,
		Y: origin.Y + ((util.Chunk2D.Y / 2) * util.Chunk2D.Y) - 1,
	}

	mgr := &RegionManager{
		changeChan:  changeChan,
		doneChan:    make(chan bool, 1),
		resetChan:   make(chan bool, 1),
		chunkStates: map[model.Vector3D]*chunkState{},
		origin:      origin,
		topLeft:     topLeft,
		bottomRight: bottomRight,
		worldMap:    m,
		scheduler:   NewScheduler(),
	}

	return mgr
}

// Start begins background routines that monitor for changes to the state of the region. When cleaning up, Stop() should
// be called to gracefully terminate this process.
func (r *RegionManager) Start() {
	go r.loop()
}

// Stop terminates scheduled events and stops the management of the map region.
func (r *RegionManager) Stop() {
	r.doneChan <- true
}

// State returns the last computed state of the region, described as a slice of response.Response messages that can
// be sent to a player's client.
func (r *RegionManager) State(trim model.Boundary) []response.Response {
	var state []response.Response
	for _, st := range r.state {
		if st.boundary&trim != 0 || st.state == nil {
			continue
		}

		state = append(state, st.state)
	}

	return state
}

// AddGroundItem adds a ground item to a tile with an optional timeout when the item should be removed.
func (r *RegionManager) AddGroundItem(itemID int, timeoutSeconds *int, globalPos model.Vector3D) {
	tile := r.worldMap.Tile(globalPos)
	if tile == nil {
		return
	}

	// add the item to the tile, and schedule a timeout to remove it if an expiration was passed
	instanceUUID := tile.AddItem(itemID)
	if timeoutSeconds != nil {
		timeout := *timeoutSeconds

		r.scheduler.Plan(&Event{
			Type:         EventRemoveExpiredGroundItem,
			Schedule:     time.Now().Add(time.Second * time.Duration(timeout)),
			InstanceUUID: instanceUUID,
			GlobalPos:    globalPos,
		})

		r.resetChan <- true
	}

	// track this change to the region state
	r.addDelta(&changeDelta{
		eventType: changeEventAddGroundItem,
		globalPos: globalPos,
		itemIDs:   []int{itemID},
	})
}

// ClearGroundItems removes all ground items on a tile.
func (r *RegionManager) ClearGroundItems(globalPos model.Vector3D) {
	tile := r.worldMap.Tile(globalPos)
	if tile == nil {
		return
	}

	existingItemIDs := tile.GroundItemIDs()
	tile.Clear()

	r.addDelta(&changeDelta{
		eventType: changeEventRemoveGroundItem,
		globalPos: globalPos,
		itemIDs:   existingItemIDs,
	})
}

// Reconcile validates the current state of the region and recomputes its state if a change has occurred. A slice of
// messages will be returned that should be dispatched to players in the region.
func (r *RegionManager) Reconcile() []response.Response {
	// perform an initial state update
	if r.state == nil {
		r.refreshRegion()
		return nil
	}

	r.mu.Lock()

	// track updates by chunk origin, relative to the region origin
	updates := map[model.Vector2D][]response.Response{}

	for _, e := range r.pendingEvents {
		chunkOrigin, relative := r.globalToChunkOriginAndRelative(e.globalPos)

		// track changes to this chunk
		chunkLocal := util.GlobalToRegionLocal(chunkOrigin).To2D()

		switch e.eventType {
		case changeEventAddGroundItem:
			// one or more ground items were added to a tile
			for _, itemID := range e.itemIDs {
				updates[chunkLocal] = append(updates[chunkLocal], response.NewShowGroundItemResponse(itemID, 1, relative))
			}

			// recompute the state of the tile the item was added to
			// TODO: can this be optimized to only update the tile itself?
			//chunkState := r.computeChunk(chunkOrigin)

			// since this chunk's state might have changed, we need to synchronize the region's memoized state
			//r.chunkStates[chunkOrigin] = chunkState
			r.syncOverallState()

		case changeEventRemoveGroundItem:
			// one or more ground items on a tile were removed
			for _, itemID := range e.itemIDs {
				updates[chunkLocal] = append(updates[chunkLocal], response.NewRemoveGroundItemResponse(itemID, relative))
			}

			// recompute the state of the tile the item was added to
			// TODO: can this be optimized to only update the tile itself?
			//chunkState := r.computeChunk(chunkOrigin)

			// since this chunk's state might have changed, we need to synchronize the region's memoized state
			//r.chunkStates[chunkOrigin] = chunkState
			r.syncOverallState()
		}
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

// loop processes events that occur within the region that affect its state.
func (r *RegionManager) loop() {
	run := true

	for run {
		select {
		case <-r.doneChan:
			// the processing loop has been shut down
			run = false

		case <-r.resetChan:
			// run another iteration since the scheduler has changed

		case <-time.After(r.scheduler.TimeUntil()):
			// handle the next scheduled event
			r.handleNextEvent()
		}
	}
}

// handleNextEvent processes the next scheduled event in this region.
func (r *RegionManager) handleNextEvent() {
	event := r.scheduler.Next()
	if event == nil {
		return
	}

	switch event.Type {
	case EventRemoveExpiredGroundItem:
		// a ground item has expired and should be removed, if it's still on a tile
		tile := r.worldMap.Tile(event.GlobalPos)
		if tile == nil {
			return
		}

		// attempt to remove the item if it still exists
		itemID := tile.RemoveItem(event.InstanceUUID)
		if itemID == nil {
			return
		}

		// track this change to the tile
		r.addDelta(&changeDelta{
			eventType: changeEventRemoveGroundItem,
			itemIDs:   []int{*itemID},
			globalPos: event.GlobalPos,
		})

		// notify on the change channel since this was an internally scheduled event
		r.changeChan <- r.origin
	default:
	}
}

// boundaryForChunk returns a bitmask describing if a chunk origin, in global coordinates, falls into one or more of
// the area padding boundaries around a region.
func (r *RegionManager) boundaryForChunk(origin model.Vector3D) model.Boundary {
	// determine if this chunk falls into the area padding around the region. we also need to add another chunk's
	// worth of padding since the client renders extra tiles further ahead of said boundary
	boundary := model.BoundaryNone
	if origin.X < r.topLeft.X-util.Chunk2D.X {
		boundary |= model.BoundaryWest
	} else if origin.X >= r.bottomRight.X+util.Chunk2D.X {
		boundary |= model.BoundaryEast
	}

	if origin.Y < r.bottomRight.Y-util.Chunk2D.X {
		boundary |= model.BoundarySouth
	} else if origin.Y >= r.topLeft.Y+util.Chunk2D.X {
		boundary |= model.BoundaryNorth
	}

	return boundary
}

// globalToChunkOriginAndRelative translates a position in global coordinates to the origin of the containing chunk,
// in global coordinates and a relative offset to that position from said origin.
func (r *RegionManager) globalToChunkOriginAndRelative(globalPos model.Vector3D) (model.Vector3D, model.Vector2D) {
	// an item was added to a tile: compute the chunk origin in global coordinates of said tile
	chunkOrigin := model.Vector3D{
		X: r.origin.X + ((globalPos.X-r.origin.X)/util.Chunk2D.X)*util.Chunk2D.X,
		Y: r.origin.Y + ((globalPos.Y-r.origin.Y)/util.Chunk2D.Y)*util.Chunk2D.Y,
		Z: r.origin.Z,
	}

	// compute the relative offsets to the tile with respect to the chunk origin
	relative := model.Vector2D{
		X: globalPos.X - chunkOrigin.X,
		Y: globalPos.Y - chunkOrigin.Y,
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
			chunkRelative := model.Vector2D{
				X: (util.Chunk2D.X * 2) + (x+util.Chunk2D.X/2)*util.Chunk2D.X,
				Y: (util.Chunk2D.Y * 2) + (y+util.Chunk2D.Y/2)*util.Chunk2D.Y,
			}

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
	boundary := r.boundaryForChunk(chunkOriginGlobal)

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
			if tileState != nil {
				batched = append(batched, tileState...)
			}
		}
	}

	if batched == nil {
		return nil
	}

	return &chunkState{
		boundary: boundary,
		state:    response.NewBatchResponse(chunkRelative, batched),
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
	for _, item := range tile.GroundItemIDs() {
		batched = append(batched, response.NewShowGroundItemResponse(item, 1, relative))
	}

	return batched
}
