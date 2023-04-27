package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
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

// RegionManager is responsible for tracking the state of a single, 2D region on the world map.
type RegionManager struct {
	changeChan    chan model.Vector3D
	doneChan      chan bool
	resetChan     chan bool
	origin        model.Vector3D
	worldMap      *model.Map
	state         []response.Response
	chunkStates   map[model.Vector3D]response.Response
	pendingEvents []*changeDelta
	scheduler     *Scheduler
}

// NewRegionManager creates a manager for a 2D region of the map. The z-coordinate will be used to determine which
// plane of the region this manager will be responsible for. A changeChan will be written to when an internal event
// causes the state of the region to change, containing the origin of the region this manager is overseeing.
func NewRegionManager(origin model.Vector3D, m *model.Map, changeChan chan model.Vector3D) *RegionManager {
	mgr := &RegionManager{
		changeChan:  changeChan,
		doneChan:    make(chan bool, 1),
		resetChan:   make(chan bool, 1),
		chunkStates: map[model.Vector3D]response.Response{},
		origin:      origin,
		worldMap:    m,
		scheduler:   NewScheduler(),
	}

	go mgr.loop()
	return mgr
}

// Stop terminates scheduled events and stops the management of the map region.
func (r *RegionManager) Stop() {
	r.doneChan <- true
}

// State returns the last computed state of the region, described as a slice of response.Response messages that can
// be sent to a player's client.
func (r *RegionManager) State() []response.Response {
	return r.state
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
	r.pendingEvents = append(r.pendingEvents, &changeDelta{
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

	r.pendingEvents = append(r.pendingEvents, &changeDelta{
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

	var updates []response.Response

	for _, e := range r.pendingEvents {
		chunkOrigin, relative := r.globalToChunkOriginAndRelative(e.globalPos)

		switch e.eventType {
		case changeEventAddGroundItem:
			// one or more ground items were added to a tile
			for _, itemID := range e.itemIDs {
				updates = append(updates, response.NewShowGroundItemResponse(itemID, 1, relative))
			}

			// recompute the state of the tile the item was added to
			// TODO: can this be optimized to only update the tile itself?
			chunkState := r.computeChunk(chunkOrigin)

			// since this chunk's state might have changed, we need to synchronize the region's memoized state
			r.chunkStates[chunkOrigin] = chunkState
			r.syncOverallState()

		case changeEventRemoveGroundItem:
			// one or more ground items on a tile were removed
			for _, itemID := range e.itemIDs {
				updates = append(updates, response.NewRemoveGroundItemResponse(itemID, relative))
			}

			// recompute the state of the tile the item was added to
			// TODO: can this be optimized to only update the tile itself?
			chunkState := r.computeChunk(chunkOrigin)

			// since this chunk's state might have changed, we need to synchronize the region's memoized state
			r.chunkStates[chunkOrigin] = chunkState
			r.syncOverallState()
		}
	}

	r.pendingEvents = nil
	return updates
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
		// FIXME: this is not reentrant
		itemID := tile.RemoveItem(event.InstanceUUID)
		if itemID == nil {
			return
		}

		// track this change to the tile
		r.pendingEvents = append(r.pendingEvents, &changeDelta{
			eventType: changeEventRemoveGroundItem,
			itemIDs:   []int{*itemID},
			globalPos: event.GlobalPos,
		})

		// notify on the change channel since this was an internally scheduled event
		r.changeChan <- r.origin
	default:
	}
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
	r.chunkStates = map[model.Vector3D]response.Response{}

	// compute batches for each chunk in this region
	for x := 0; x < util.Region3D.X*util.Chunk2D.X; x += util.Chunk2D.X {
		for y := 0; y < util.Region3D.Y*util.Chunk2D.Y; y += util.Chunk2D.Y {
			chunkOrigin := model.Vector3D{
				X: r.origin.X + x,
				Y: r.origin.Y + y,
				Z: r.origin.Z,
			}

			// compute the state of this chunk and add it to the overall region state
			chunkState := r.computeChunk(chunkOrigin)
			if chunkState == nil {
				delete(r.chunkStates, chunkOrigin)
			} else {
				r.chunkStates[chunkOrigin] = chunkState
				r.state = append(r.state, chunkState)
			}
		}
	}
}

// computeChunk builds a response.Response slice containing the current state of a chunk in a region. The
// chunkOriginGlobal should be the origin of the chunk in global coordinates. If there are no updates for this chunk,
// then a nil will be returned instead.
func (r *RegionManager) computeChunk(chunkOriginGlobal model.Vector3D) response.Response {
	var batched []response.Response
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

			relative := model.Vector2D{
				X: x,
				Y: y,
			}

			// compute the state of the tile
			tileState := r.computeTile(tilePos, relative)
			if tileState != nil {
				batched = append(batched, tileState...)
			}
		}
	}

	if batched == nil {
		return nil
	}

	return response.NewBatchResponse(origin.To2D(), batched)
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
