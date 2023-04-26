package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/util"
	"io"
)

// mapObject is an object that is located on the map.
type mapObject struct {
	ID          int
	Position    model.Vector3D
	ObjectType  int
	Orientation int
}

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

	objectCache := map[int][]*mapObject{}
	m := model.NewMap()
	maxX, maxY := 0, 0

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

		// read terrain data for this region
		terrainID := terrainIndices[i]
		err = l.readTerrainArea(terrainID, global, m)
		if err != nil {
			return nil, err
		}

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
		zValues := map[int]bool{}
		for _, obj := range regionObjects {
			// get the tile at this location, or create one if this is the first time we've encountered this location
			tilePos := obj.Position.Add(global)
			tile := m.Tile(tilePos)
			if tile == nil {
				m.SetTile(tilePos, &model.Tile{})
			}

			// keep track of the maximum x- and y- coordinates on the map
			maxX = util.Max(maxX, tilePos.X)
			maxY = util.Max(maxY, tilePos.Y)

			// mark that an object exists on this z-coordinate
			zValues[tilePos.Z] = true

			// look up the object by its id
			object := objects[obj.ID]
			if object == nil {
				return nil, fmt.Errorf("invalid object ID at %v: %d", tilePos, obj.ID)
			}

			tile.AddObject(object)
		}

		// add this region (accounting only for z-coordinates with data) to the map's known region origins
		for z, _ := range zValues {
			m.RegionOrigins = append(m.RegionOrigins, model.Vector3D{
				X: global.X,
				Y: global.Y,
				Z: z,
			})
		}
	}

	m.MaxTile = model.Vector2D{
		X: maxX,
		Y: maxY,
	}

	return m, nil
}

// reader returns a DataReader for terrain or object data identified by an ID.
func (l *MapLoader) reader(id int) (*DataReader, error) {
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
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	return NewDataReader(data), nil
}

// readTerrainArea loads map terrain data for a region into a map.
func (l *MapLoader) readTerrainArea(id int, regionGlobal model.Vector3D, m *model.Map) error {
	r, err := l.reader(id)
	if err != nil {
		return err
	}

	// read terrain data for the entire prism encompassed by this region
	for z := 0; z < 4; z++ {
		for x := 0; x < util.Region3D.X; x++ {
			for y := 0; y < util.Region3D.Y; y++ {
				var below *model.Tile
				if z > 0 {
					below = m.Tile(model.Vector3D{
						X: regionGlobal.X + x,
						Y: regionGlobal.Y + y,
						Z: z - 1,
					})
				}

				tile, err := l.readTerrainTile(r, below)
				m.SetTile(model.Vector3D{
					X: regionGlobal.X + x,
					Y: regionGlobal.Y + y,
					Z: z,
				}, tile)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// readTerrainTile loads a map tile.
func (l *MapLoader) readTerrainTile(r *DataReader, below *model.Tile) (*model.Tile, error) {
	tile := &model.Tile{}
	hasMoreAttributes := true

	for hasMoreAttributes {
		// read 1 byte for the op code
		op, err := r.Byte()
		if err != nil {
			return nil, err
		}

		switch op {
		case 0x00:
			// no explicit height; calculate it based on the tile's location
			if below == nil {
				// TODO
				tile.Height = 0
			} else {
				tile.Height = below.Height - 0xF0
			}

			hasMoreAttributes = false

		case 0x01:
			// read 1 byte for the vertex height
			height, err := r.Byte()
			if err != nil {
				return nil, err
			}

			// adjust the height
			if height == 1 {
				height = 0
			}

			// if there is no tile below this one, set the height explicitly otherwise derive it from the one below
			if below == nil {
				tile.Height = int(height) * -8
			} else {
				tile.Height = below.Height - (int(height) * 8)
			}

			hasMoreAttributes = false

		default:
			if op <= 0x31 {
				// read 1 byte for the overlay ID
				overlayID, err := r.Byte()
				if err != nil {
					return nil, err
				}

				tile.OverlayID = int(overlayID)
			} else if op <= 0x51 {
				tile.RenderFlag = int(op - 0x31)
			} else {
				tile.UnderlayID = int(op - 0x51)
			}
		}
	}

	return tile, nil
}

// readObjects loads map object data.
func (l *MapLoader) readObjects(id int) ([]*mapObject, error) {
	r, err := l.reader(id)
	if err != nil {
		return nil, err
	}

	var objects []*mapObject
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

			objects = append(objects, &mapObject{
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
