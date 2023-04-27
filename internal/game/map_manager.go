package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
	"sync"
)

// MapManager is responsible for managing the state of the entire world map.
type MapManager struct {
	changeChan     chan model.Vector3D
	doneChan       chan bool
	regions        map[model.Vector3D]*RegionManager
	pendingRegions map[model.Vector3D]bool
	worldMap       *model.Map
}

// NewMapManager creates a new manager for a world map.
func NewMapManager(m *model.Map) *MapManager {
	regions := map[model.Vector3D]*RegionManager{}
	changeChan := make(chan model.Vector3D, len(m.RegionOrigins))

	for _, origin := range m.RegionOrigins {
		regions[origin] = NewRegionManager(origin, m, changeChan)
	}

	mgr := &MapManager{
		changeChan:     changeChan,
		doneChan:       make(chan bool, 1),
		pendingRegions: map[model.Vector3D]bool{},
		regions:        regions,
		worldMap:       m,
	}

	go mgr.loop()
	return mgr
}

// Stop terminates scheduled events and stops the management of the map.
func (m *MapManager) Stop() {
	m.doneChan <- true

	for _, region := range m.regions {
		region.Stop()
	}
}

// State returns the last computed state of a 2D region. The origin should be the region origin in global coordinates,
// and the z-coordinate will be used to determine which plane of a region to return. If no region exists at this origin,
// nil will be returned instead.
func (m *MapManager) State(origin model.Vector3D) []response.Response {
	region, ok := m.regions[origin]
	if !ok {
		return nil
	}

	return region.State()
}

// AddGroundItem adds a ground item to the top of a tile with an optional timeout (in seconds) when that item should
// automatically be removed.
func (m *MapManager) AddGroundItem(itemID int, timeoutSeconds *int, globalPos model.Vector3D) {
	region := util.GlobalToRegionGlobal(globalPos)

	mgr, ok := m.regions[region]
	if !ok {
		return
	}

	mgr.AddGroundItem(itemID, timeoutSeconds, globalPos)
	m.pendingRegions[region] = true
}

func (m *MapManager) ClearGroundItems(globalPos model.Vector3D) {
	region := util.GlobalToRegionGlobal(globalPos)

	mgr, ok := m.regions[region]
	if !ok {
		return
	}

	mgr.ClearGroundItems(globalPos)
	m.pendingRegions[region] = true
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

	// process each region that has pending updates available
	for origin, _ := range m.pendingRegions {
		updates[origin] = m.regions[origin].Reconcile()
	}

	m.pendingRegions = map[model.Vector3D]bool{}
	return updates
}

// loop processes state changes to the map that occur internally.
func (m *MapManager) loop() {
	run := true

	for run {
		select {
		case <-m.doneChan:
			// the processing loop has been shut down
			run = false

		case region := <-m.changeChan:
			// a region's state has changed internally; track it for reconciliation
			// FIXME: this is not reentrant
			m.pendingRegions[region] = true
		}
	}
}
