package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/util"
	"io"
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

func (l *MapLoader) Load(objects []*model.WorldObject) (*model.Map, error) {
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
	m := model.NewMap()

	// load map data for each region
	for i := 0; i < len(coordinateIndices); i++ {
		// unpack the region coordinates for this index
		regionX := coordinateIndices[i] >> 8
		regionY := coordinateIndices[i] & 0xFF

		// convert the region origin to global coordinates
		global := model.Vector3D{
			X: regionX * util.Region3D.X,
			Y: regionY * util.Region3D.Y,
			Z: 0,
		}

		// initialize tiles in this region
		for z := 0; z <= 4; z++ {
			for x := global.X; x < global.X+util.Region3D.X; x++ {
				for y := global.Y; y < global.Y+util.Region3D.Y; y++ {
					m.PutTile(model.Vector3D{
						X: x,
						Y: y,
						Z: z,
					})
				}
			}
		}

		// TODO: terrain data
		_ = terrainIndices[i]

		// read the map objects that are location on this region
		objectsID := objectIndices[i]
		regionObjects, ok := objectCache[objectsID]
		if !ok {
			regionObjects, err = l.readObjects(objectsID)
			if err != nil {
				return nil, err
			}
		}

		// connect the object ids to their objects and place them on tiles
		for _, obj := range regionObjects {
			tilePos := obj.Position.Add(global)
			tile := m.Tile(tilePos)

			// look up the object by its id
			object := objects[obj.ID]
			if object == nil {
				return nil, fmt.Errorf("invalid object ID at %v: %d", tilePos, obj.ID)
			}

			tile.AddObject(object)
		}
	}

	return m, nil
}

// readObjects loads map object data.
func (l *MapLoader) readObjects(id int) ([]*model.MapObject, error) {
	compressed, err := l.cache.Data(id)
	if err != nil {
		return nil, err
	}

	// decompress the data
	gzipReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	data, err := io.ReadAll(gzipReader)
	if err != io.ErrUnexpectedEOF {
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
