package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
)

// RegionManager is responsible for tracking the state of a single, 2D region on the world map.
type RegionManager struct {
	origin   model.Vector3D
	worldMap *model.Map
	state    []response.Response
}

// NewRegionManager creates a manager for a 2D region of the map. The z-coordinate will be used to determine which
// plane of the region this manager will be responsible for.
func NewRegionManager(origin model.Vector3D, m *model.Map) *RegionManager {
	return &RegionManager{
		origin:   origin,
		worldMap: m,
	}
}

// State returns the last computed state of the region, described as a slice of response.Response messages that can
// be sent to a player's client.
func (r *RegionManager) State() []response.Response {
	return r.state
}

// Reconcile validates the current state of the region and recomputes its state if a change has occurred.
func (r *RegionManager) Reconcile() {
	// perform an initial state update
	if r.state == nil {
		r.state = r.computeRegion()
		return
	}
}

// computeRegion builds a slice of BatchResponse messages describing the current state of a map region. The origin
// should be the region origin in global coordinates.
func (r *RegionManager) computeRegion() []response.Response {
	var batches []response.Response

	// compute batches for each chunk in this region
	for x := 0; x < util.Region3D.X*util.Chunk2D.X; x += util.Chunk2D.X {
		for y := 0; y < util.Region3D.Y*util.Chunk2D.Y; y += util.Chunk2D.Y {
			chunkBatches := r.computeChunk(model.Vector3D{
				X: r.origin.X + x,
				Y: r.origin.Y + y,
				Z: r.origin.Z,
			})

			if chunkBatches != nil {
				batches = append(batches, chunkBatches)
			}
		}
	}

	return batches
}

// computeChunk builds a BatchResponse containing the current state of a chunk in a region. The chunkOriginGlobal
// should be the origin of the chunk in global coordinates. If there are no updates for this chunk, then a nil will
// be returned instead.
func (r *RegionManager) computeChunk(chunkOriginGlobal model.Vector3D) *response.BatchResponse {
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
