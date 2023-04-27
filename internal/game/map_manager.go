package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
	"sync"
)

// MapManager is responsible for managing the state of the entire world map.
type MapManager struct {
	regions        map[model.Vector3D]*RegionManager
	pendingRegions map[model.Vector3D]bool
	worldMap       *model.Map
}

// NewMapManager creates a new manager for a world map.
func NewMapManager(m *model.Map) *MapManager {
	regions := map[model.Vector3D]*RegionManager{}

	for _, origin := range m.RegionOrigins {
		regions[origin] = NewRegionManager(origin, m)
	}

	return &MapManager{
		pendingRegions: map[model.Vector3D]bool{},
		regions:        regions,
		worldMap:       m,
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

func (m *MapManager) AddGroundItem(itemID int, globalPos model.Vector3D) {
	region := util.GlobalToRegionGlobal(globalPos)

	mgr, ok := m.regions[region]
	if !ok {
		return
	}

	mgr.AddGroundItem(itemID, globalPos)
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
