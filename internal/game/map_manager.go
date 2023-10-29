package game

import (
	"github.com/google/uuid"
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
	changeEventUpdateGroundItem
)

// changeDeltaItem is an Item that was added or removed on a tile.
type changeDeltaItem struct {
	itemID    int
	amount    int
	oldAmount int
}

// changeDelta is a mutation to a tile that should be tracked.
type changeDelta struct {
	eventType changeEventType
	globalPos model.Vector3D
	items     []changeDeltaItem
}

// MapManager is responsible for managing the state of the entire world map.
type MapManager struct {
	// doneChan is a channel that tracks if the manager should terminate its internal goroutines.
	doneChan chan bool
	// changeChan is a channel that forces the manager to rerun its internal goroutines.
	changeChan chan bool
	// regions is a map of region origins, in global coordinates, to their managers.
	regions map[model.Vector3D]*RegionManager
	// pendingRegions is a map of region origins, in global coordinates, to flags if they need to be reconciled.
	pendingRegions map[model.Vector3D]bool
	// scheduler is used to track events on the map.
	scheduler *Scheduler
	// worldMap is a pointer to the model.Map that this MapManager is responsible for managing.
	worldMap *model.Map
	// mu is used to synchronize access to volatile fields.
	mu sync.Mutex
}

// NewMapManager creates a new manager for a world map.
func NewMapManager(m *model.Map) *MapManager {
	regions := map[model.Vector3D]*RegionManager{}

	// create a region manager for each known region. because each RegionManager spans an area that overlaps with the
	// regions directly bordering it, certain tiles will fall under the purview of multiple managers.
	for _, origin := range m.RegionOrigins {
		mgr := NewRegionManager(origin, m)
		regions[origin] = mgr
	}

	changeChan := make(chan bool, 1)

	mgr := &MapManager{
		doneChan:       make(chan bool, 1),
		changeChan:     changeChan,
		pendingRegions: map[model.Vector3D]bool{},
		regions:        regions,
		scheduler:      NewScheduler(changeChan),
		worldMap:       m,
	}

	return mgr
}

// Start begins background routines that monitor for changes to the state of the map. When cleaning up, Stop() should be
// called to gracefully terminate this process.
func (m *MapManager) Start() {
	go m.loop()
}

// Stop terminates scheduled events and stops the management of the map.
func (m *MapManager) Stop() {
	m.doneChan <- true
}

// State returns the last computed state of a 2D region. The origin should be the region origin in global coordinates,
// and the z-coordinate will be used to determine which plane of a region to return. If no region exists at this origin,
// nil will be returned instead.
func (m *MapManager) State(origin model.Vector3D, trim model.Boundary) []response.Response {
	region, ok := m.regions[origin]
	if !ok {
		return nil
	}

	return region.State(trim)
}

// AddPlayer adds a player to the world map at the region whose coordinates correspond to regionGlobal.
func (m *MapManager) AddPlayer(pe *playerEntity, regionGlobal model.Vector3D) {
	regions := m.findOverlappingRegions(regionGlobal)
	for _, origin := range regions {
		region := m.regions[origin]
		region.AddPlayer(pe)
	}
}

// RemovePlayer removes a player from the world map from the region specified by regionGlobal.
func (m *MapManager) RemovePlayer(pe *playerEntity, regionGlobal model.Vector3D) {
	regions := m.findOverlappingRegions(regionGlobal)
	for _, origin := range regions {
		region := m.regions[origin]
		region.RemovePlayer(pe)
	}
}

// AddNPC adds an NPC to the world map at the region whose coordinates correspond to regionGlobal.
func (m *MapManager) AddNPC(ne *npcEntity, regionGlobal model.Vector3D) {
	regions := m.findOverlappingRegions(regionGlobal)
	for _, origin := range regions {
		region := m.regions[origin]
		region.AddNPC(ne)
	}
}

// RemoveNPC removes an NPC from the world map from the region specified by regionGlobal.
func (m *MapManager) RemoveNPC(ne *npcEntity, regionGlobal model.Vector3D) {
	regions := m.findOverlappingRegions(regionGlobal)
	for _, origin := range regions {
		region := m.regions[origin]
		region.RemoveNPC(ne)
	}
}

// FindSpectators returns a map of player IDs to playerEntity instances that are within viewable distance of the given
// player.
func (m *MapManager) FindSpectators(pe *playerEntity) map[int]*playerEntity {
	spectators := map[int]*playerEntity{}

	regions := m.findOverlappingRegions(pe.player.GlobalPos)
	for _, origin := range regions {
		region := m.regions[origin]

		for _, pe := range region.FindSpectators(pe) {
			spectators[pe.player.ID] = pe
		}
	}

	return spectators
}

// AddGroundItem adds a ground Item to the top of a tile with an optional timeout (in seconds) when that Item should
// automatically be removed. Stackable items will be added to an existing stackable with the same Item ID, if one
// exists, or they will be placed as new items on the tile.
func (m *MapManager) AddGroundItem(itemID, amount int, stackable bool, timeoutSeconds *int, globalPos model.Vector3D) {
	tile := m.worldMap.Tile(globalPos)
	if tile == nil {
		return
	}

	var instanceUUID uuid.UUID
	newlyAdded := true
	oldAmount := 0

	// add the Item to the tile. if the Item is stackable, attempt to find an update an newlyAdded stackable with the
	// same Item id
	if stackable {
		instanceUUID, newlyAdded, oldAmount = tile.AddStackableItem(itemID, amount)
	} else {
		instanceUUID = tile.AddItem(itemID)
	}

	// find each region manager that is aware of this tile and inform them about the change
	regions := m.findOverlappingRegions(globalPos)
	for _, origin := range regions {
		region := m.regions[origin]

		if newlyAdded {
			region.MarkGroundItemAdded(itemID, amount, globalPos)
		} else {
			region.MarkGroundItemUpdated(itemID, oldAmount, amount+oldAmount, globalPos)
		}

		m.addPendingRegion(origin)
	}

	// if this Item has an expiration, schedule an event to remove it after the fact
	if timeoutSeconds != nil {
		timeout := *timeoutSeconds
		m.scheduler.Plan(&Event{
			Type:         EventRemoveExpiredGroundItem,
			Schedule:     time.Now().Add(time.Second * time.Duration(timeout)),
			InstanceUUID: instanceUUID,
			GlobalPos:    globalPos,
		})
	}
}

// RemoveGroundItem attempts to remove a ground Item with the given ID at a position, in global coordinates. If the Item
// was found and removed, a pointer to its model.TileGroundItem model will be returned.
func (m *MapManager) RemoveGroundItem(itemID int, globalPos model.Vector3D) *model.TileGroundItem {
	tile := m.worldMap.Tile(globalPos)
	if tile == nil {
		return nil
	}

	// attempt to remove the ground Item, if it still exists on this tile
	item := tile.RemoveItemByID(itemID)
	if item == nil {
		return nil
	}

	// find each region manager that is aware of this tile and inform them about the change
	regions := m.findOverlappingRegions(globalPos)
	for _, origin := range regions {
		region := m.regions[origin]

		region.MarkGroundItemsCleared([]int{itemID}, globalPos)
		m.addPendingRegion(origin)
	}

	return item
}

// ClearGroundItems removes all ground items on a tile.
func (m *MapManager) ClearGroundItems(globalPos model.Vector3D) {
	tile := m.worldMap.Tile(globalPos)
	if tile == nil {
		return
	}

	items := tile.GroundItems()
	tile.Clear()

	itemIDs := make([]int, len(items))
	for i, item := range items {
		itemIDs[i] = item.ItemID
	}

	// find each region manager that is aware of this tile and inform them about the change
	regions := m.findOverlappingRegions(globalPos)
	for _, origin := range regions {
		region := m.regions[origin]

		region.MarkGroundItemsCleared(itemIDs, globalPos)
		m.addPendingRegion(origin)
	}
}

// WarmUp computes the initial state of the world map. This should generally be called only once before the game state
// begins changing.
func (m *MapManager) WarmUp() {
	var wg sync.WaitGroup

	// recompute each region's state in isolation
	for _, mgr := range m.regions {
		wg.Add(1)

		go func(mgr *RegionManager) {
			defer wg.Done()
			mgr.Reconcile()
		}(mgr)
	}

	wg.Wait()
}

// Reconcile validates the current state of the entire world map and recomputes its state if a change has occurred.
func (m *MapManager) Reconcile() map[model.Vector3D][]response.Response {
	updates := map[model.Vector3D][]response.Response{}

	m.mu.Lock()
	defer m.mu.Unlock()

	// process each region that has pending updates available
	for origin, _ := range m.pendingRegions {
		updates[origin] = m.regions[origin].Reconcile()
	}

	m.pendingRegions = map[model.Vector3D]bool{}
	return updates
}

// findOverlappingRegions returns a slice of region origins, in global coordinates, for each region that manages a
// position, in global coordinates. Since map regions span an area that overlaps with neighboring regions, a single
// position may be managed by multiple regions. This method will compute each of those overlapping regions, including
// the region which actually contains the position itself.
func (m *MapManager) findOverlappingRegions(globalPos model.Vector3D) []model.Vector3D {
	baseRegion := util.GlobalToRegionOrigin(globalPos)
	allRegions := []model.Vector3D{util.GlobalToRegionGlobal(globalPos)}

	checkAndAddRegion := func(x, y int) {
		neighborOrigin := util.RegionOriginToGlobal(model.Vector2D{
			X: baseRegion.X + x*util.Chunk2D.X,
			Y: baseRegion.Y + y*util.Chunk2D.X,
		})
		neighborOrigin.Z = globalPos.Z

		region, ok := m.regions[neighborOrigin]
		if !ok {
			return
		}

		if region.Contains(globalPos) {
			allRegions = append(allRegions, neighborOrigin)
		}
	}

	// check if all regions in each cardinal direction also manages this position
	checkAndAddRegion(-1, 0)
	checkAndAddRegion(1, 0)
	checkAndAddRegion(0, 1)
	checkAndAddRegion(0, -1)

	return allRegions
}

// addPendingRegion marks a region that has had at least one change and should be reported the next time the manager
// reconciles its changes. The origin should be the region origin in global coordinates.
func (m *MapManager) addPendingRegion(origin model.Vector3D) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pendingRegions[origin] = true
}

// loop processes state changes to the map that occur internally.
func (m *MapManager) loop() {
	run := true

	for run {
		select {
		case <-m.doneChan:
			// the processing loop has been shut down
			run = false

		case <-m.changeChan:
			// run another iteration since the scheduler has changed

		case <-time.After(m.scheduler.TimeUntil()):
			// handle the next scheduled event
			m.handleNextEvent()
		}
	}
}

// handleNextEvent processes the next scheduled event on the map.
func (m *MapManager) handleNextEvent() {
	event := m.scheduler.Next()
	if event == nil {
		return
	}

	switch event.Type {
	case EventRemoveExpiredGroundItem:
		// a ground Item has expired and should be removed, if it's still on a tile
		tile := m.worldMap.Tile(event.GlobalPos)
		if tile == nil {
			return
		}

		// attempt to remove the Item if it still exists
		itemID := tile.RemoveItemByInstanceUUID(event.InstanceUUID)
		if itemID == nil {
			return
		}

		// find each region manager that is aware of this tile and inform them about the change
		regions := m.findOverlappingRegions(event.GlobalPos)
		for _, origin := range regions {
			region := m.regions[origin]

			region.MarkGroundItemsCleared([]int{*itemID}, event.GlobalPos)
			m.addPendingRegion(origin)
		}
	default:
	}
}
