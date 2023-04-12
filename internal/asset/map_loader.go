package asset

import (
	"github.com/mbpolan/openmcs/internal/model"
)

// MapLoader loads world map data from game asset files.
type MapLoader struct {
	archive *Archive
	cache   *CacheFile
}

func NewMapLoader(archive *Archive, cache *CacheFile) *MapLoader {
	return &MapLoader{
		archive: archive,
		cache:   cache,
	}
}

func (l *MapLoader) Load() (*model.Map, error) {
	// extract the map index file
	mapIndex, err := l.archive.File("map_index")
	if err != nil {
		return nil, err
	}

	r := NewDataReader(mapIndex)

	// compute how many regions are in the map
	numRegions := len(mapIndex) / 7
	coordinateIndices := make([]int, numRegions)
	terrainIndices := make([]int, numRegions)
	objectIndices := make([]int, numRegions)
	preloadRegions := make([]bool, numRegions)

	// read coordinate, terrain and object indices for the entire world map
	// additionally read a byte that indicates if a map region should be preloaded by the client
	for i := 0; i < numRegions; i++ {
		v, err := r.Uint16()
		if err != nil {
			return nil, err
		}

		coordinateIndices[i] = int(v)

		v, err = r.Uint16()
		if err != nil {
			return nil, err
		}

		terrainIndices[i] = int(v)

		v, err = r.Uint16()
		if err != nil {
			return nil, err
		}

		objectIndices[i] = int(v)

		b, err := r.Byte()
		if err != nil {
			return nil, err
		}

		preloadRegions[i] = b == 0x01
	}

	objectCache := map[int][]*model.MapObject{}

	// load map data for each region
	regions := map[int]map[int]*model.MapRegion{}
	for i := 0; i < len(coordinateIndices); i++ {
		// unpack the region coordinates for this index
		x := coordinateIndices[i] >> 8
		y := coordinateIndices[i] & 0xFF

		if regions[x] == nil {
			regions[x] = map[int]*model.MapRegion{}
		}

		// read the map objects that are location on this region
		objectsID := objectIndices[i]
		objects, ok := objectCache[objectsID]
		if !ok {
			objects, err = l.readObjects(objectsID)
			if err != nil {
				return nil, err
			}
		}

		regions[x][y] = &model.MapRegion{
			TerrainID: terrainIndices[i],
			Objects:   objects,
		}
	}

	return &model.Map{
		Regions: regions,
	}, nil
}

// readObjects loads map object data.
func (l *MapLoader) readObjects(id int) ([]*model.MapObject, error) {
	data, err := l.cache.Data(id)
	if err != nil {
		return nil, err
	}

	r := NewDataReader(data)

	var objects []*model.MapObject
	hasMoreObjects := true
	objectID := -1

	for hasMoreObjects {
		// read the object id
		idOffset, err := r.VarByte()
		if err != nil {
			return nil, err
		}

		// no more objects to read
		if idOffset == 0 {
			hasMoreObjects = false
			break
		}

		objectID += int(idOffset)
		posOffset := 0

		// reach each instance of the object in this area
		hasMorePositions := true
		for hasMorePositions {
			offset, err := r.VarByte()
			if err != nil {
				return nil, err
			}

			// no more position data to read
			if offset == 0 {
				hasMorePositions = false
				break
			}

			attributes, err := r.Byte()
			if err != nil {
				return nil, err
			}

			// read the position of this object relative to the origin of the map area
			posOffset += int(offset) - 1
			tileX := (posOffset >> 6) & 0x3F
			tileY := posOffset & 0x3F
			tileZ := posOffset >> 12

			// extract additional attributes about the object
			objType := attributes >> 2
			objOrient := attributes & 0x03

			objects = append(objects, &model.MapObject{
				ID: objectID,
				Position: model.Vector3D{
					X: tileX,
					Y: tileY,
					Z: tileZ,
				},
				ObjectType:  int(objType),
				Orientation: int(objOrient),
			})
		}
	}

	return objects, nil
}
