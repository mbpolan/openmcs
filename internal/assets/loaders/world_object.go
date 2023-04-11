package loaders

import (
	"github.com/mbpolan/openmcs/internal/assets"
	"github.com/mbpolan/openmcs/internal/models"
	"io"
)

const (
	opEndDefinition    byte = 0x00
	opModelIDsTypes         = 0x01
	opName                  = 0x02
	opDescription           = 0x03
	opModelIDs              = 0x05
	opSizeX                 = 0x0E
	opSizeY                 = 0x0F
	opNotSolidFlag          = 0x11
	opNotWalkableFlag       = 0x12
	opActions               = 0x13
	opTerrainFlag           = 0x15
	opShadingFlag           = 0x16
	opWallFlag              = 0x17
	opAnimationID           = 0x18
	opOffsetAmpFlag         = 0x1C
	opAmbientFlag           = 0x1D
	opActionListStart       = 0x1E
	opActionListEnd         = 0x26
	opDiffuseFlag           = 0x27
	opColorCount            = 0x28
	opIcon                  = 0x3C
	opRotatedFlag           = 0x3E
	opShadowFlag            = 0x40
	opScaleX                = 0x41
	opScaleY                = 0x42
	opScaleZ                = 0x43
	opMapScene              = 0x44
	opFace                  = 0x45
	opTranslateX            = 0x46
	opTranslateY            = 0x47
	opTranslateZ            = 0x48
	opUnknown1              = 0x49
	opUnwalkableFlag        = 0x4A
	opStaticFlag            = 0x4B
	opEndAttributeList      = 0x4D
)

// WorldObjectLoader loads world object data from game asset files.
type WorldObjectLoader struct {
	archive *assets.Archive
}

func NewWorldObjectLoader(archive *assets.Archive) *WorldObjectLoader {
	return &WorldObjectLoader{
		archive: archive,
	}
}

func (l *WorldObjectLoader) Load() ([]*models.WorldObject, error) {
	// extract the files containing object data
	dataFile, err := l.archive.File("loc.dat")
	if err != nil {
		return nil, err
	}

	idxFile, err := l.archive.File("loc.idx")
	if err != nil {
		return nil, err
	}

	dataReader := assets.NewDataReader(dataFile)
	idxReader := assets.NewDataReader(idxFile)

	// read the number of objects in the file
	numObjects, err := idxReader.Uint16()
	if err != nil {
		return nil, err
	}

	objects := make([]*models.WorldObject, numObjects)

	// read each object definition from the data file
	offset := 2
	for i := 0; i < int(numObjects); i++ {
		_, err := dataReader.Seek(int64(offset), io.SeekStart)
		if err != nil {
			return nil, err
		}

		object, err := l.readObject(i, dataReader)
		if err != nil {
			return nil, err
		}

		objects[i] = object

		// move to the next offset
		nextOffset, err := idxReader.Uint16()
		if err != nil {
			return nil, err
		}

		offset += int(nextOffset)
	}

	return objects, nil
}

func (l *WorldObjectLoader) readObject(id int, r *assets.DataReader) (*models.WorldObject, error) {
	object := &models.WorldObject{
		ID: id,
	}

	hasMoreAttributes := true
	for hasMoreAttributes {
		// read the next op code to determine what attribute follows
		op, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		switch op {
		case opEndDefinition:
			// finished reading object definition
			hasMoreAttributes = false

		case opModelIDsTypes:
			// read number of models
			numModels, err := r.Byte()
			if err != nil {
				return nil, err
			}

			_, err = r.Seek(int64(numModels*3), io.SeekCurrent)
			if err != nil {
				return nil, err
			}

		case opName:
			// read the object's name
			name, err := r.String()
			if err != nil {
				return nil, err
			}

			object.Name = name

		case opDescription:
			// read the object's description
			desc, err := r.String()
			if err != nil {
				return nil, err
			}

			object.Description = desc

		case opModelIDs:
			// read number of model ids
			numModelIDs, err := r.Byte()
			if err != nil {
				return nil, err
			}

			_, err = r.Seek(int64(numModelIDs)*2, io.SeekCurrent)
			if err != nil {
				return nil, err
			}

		case opSizeX:
			// read the object's size along the x-axis
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.Size.X = int(b)

		case opSizeY:
			// read the object's size along the y-axis
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.Size.Y = int(b)

		case opNotSolidFlag:
			// flag the object as solid
			object.Solid = false

		case opNotWalkableFlag:
			// flag that the object cannot be walked on
			object.Walkable = false

		case opActions:
			// read flag indicating if object has actions
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.HasActions = b == 0x01

		case opTerrainFlag:
			object.AdjustToTerrain = true

		case opShadingFlag:
			object.DelayedShading = true

		case opWallFlag:
			object.Wall = true

		case opAnimationID:
			// skip 2 bytes containing animation id
			_, err := r.Uint16()
			if err != nil {
				return nil, err
			}

		case opOffsetAmpFlag:
			// read flag indicating if object has its offset amplified
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.OffsetAmplified = b == 0x01

		case opAmbientFlag:
			// skip byte indicating if object has ambience
			_, err := r.Byte()
			if err != nil {
				return nil, err
			}

		case opDiffuseFlag:
			// skip byte indicating if object has diffuse properties
			_, err := r.Byte()
			if err != nil {
				return nil, err
			}

		case opColorCount:
			// read byte containing number of colors
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			// there are two, two-byte values containing the modified and original color
			_, err = r.Seek(int64(b)*4, io.SeekCurrent)
			if err != nil {
				return nil, err
			}

		case opIcon:
			// skip 2 bytes containing object icon
			_, err := r.Uint16()
			if err != nil {
				return nil, err
			}

		case opRotatedFlag:
			object.Rotated = true

		case opShadowFlag:
			object.Shadowless = true

		case opScaleX:
			// read object's scale along x-axis
			scaleX, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Scale.X = int(scaleX)

		case opScaleY:
			// read object's scale along y-axis
			scaleY, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Scale.Y = int(scaleY)

		case opScaleZ:
			// read object's scale along z-axis
			scaleZ, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Scale.Z = int(scaleZ)

		case opMapScene:
			// read object's map scene id
			mapSceneID, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.MapSceneID = int(mapSceneID)

		case opFace:
			// read object's face id
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.FaceID = int(b)

		case opTranslateX:
			// read object's translation along x-axis
			translateX, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Translation.X = int(translateX)

		case opTranslateY:
			// read object's translation along y-axis
			translateY, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Translation.Y = int(translateY)

		case opTranslateZ:
			// read object's translation along z-axis
			translateZ, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Translation.Z = int(translateZ)

		case opUnknown1:
			// unknown attribute
			break

		case opUnwalkableFlag:
			object.UnwalkableSolid = true

		case opStaticFlag:
			// read flag indicating if object is static
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.Static = b == 0x01

		case opEndAttributeList:
			// end of attribute list, continue reading additional object data
			hasMoreAttributes = false

			// unknown property?
			variableID, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			if variableID != 0xFFFF {
				object.VariableID = int(variableID)
			}

			// read configuration id for the object
			configID, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			if configID != 0xFFFF {
				object.ConfigID = int(configID)
			}

			// read ids for children objects
			numChildren, err := r.Byte()
			object.ChildIDs = make([]int, int(numChildren)+1)
			for i := 0; i <= int(numChildren); i++ {
				childID, err := r.Uint16()
				if err != nil {
					return nil, err
				}

				if childID != 0xFFFF {
					object.ChildIDs[i] = int(childID)
				} else {
					object.ChildIDs[i] = -1
				}
			}

		default:
			if op >= opActionListStart && op <= opActionListEnd {
				// read actions for this object
				if len(object.Actions) == 0 {
					object.Actions = make([]string, 5)
				}

				// read action lists
				action, err := r.String()
				if err != nil {
					return nil, err
				}

				// ignore special/hidden actions
				if action != "hidden" {
					idx := int(op - opActionListStart)
					object.Actions[idx] = action
				}
			}
		}
	}

	return object, nil
}
