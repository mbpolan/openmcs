package loader

import (
	"github.com/mbpolan/openmcs/internal/asset"
	"github.com/mbpolan/openmcs/internal/model"
	"io"
)

const (
	opObjectEndDefinition    byte = 0x00
	opObjectModelIDsTypes         = 0x01
	opObjectName                  = 0x02
	opObjectDescription           = 0x03
	opObjectModelIDs              = 0x05
	opObjectSizeX                 = 0x0E
	opObjectSizeY                 = 0x0F
	opObjectNotSolidFlag          = 0x11
	opObjectNotWalkableFlag       = 0x12
	opObjectActions               = 0x13
	opObjectTerrainFlag           = 0x15
	opObjectShadingFlag           = 0x16
	opObjectWallFlag              = 0x17
	opObjectAnimationID           = 0x18
	opObjectOffsetAmpFlag         = 0x1C
	opObjectAmbientFlag           = 0x1D
	opObjectActionListStart       = 0x1E
	opObjectActionListEnd         = 0x26
	opObjectDiffuseFlag           = 0x27
	opObjectColorCount            = 0x28
	opObjectIcon                  = 0x3C
	opObjectRotatedFlag           = 0x3E
	opObjectShadowFlag            = 0x40
	opObjectScaleX                = 0x41
	opObjectScaleY                = 0x42
	opObjectScaleZ                = 0x43
	opObjectMapScene              = 0x44
	opObjectFace                  = 0x45
	opObjectTranslateX            = 0x46
	opObjectTranslateY            = 0x47
	opObjectTranslateZ            = 0x48
	opObjectUnknown1              = 0x49
	opObjectUnwalkableFlag        = 0x4A
	opObjectStaticFlag            = 0x4B
	opObjectEndAttributeList      = 0x4D
)

// WorldObjectLoader loads world object data from game asset files.
type WorldObjectLoader struct {
	archive *asset.Archive
}

func NewWorldObjectLoader(archive *asset.Archive) *WorldObjectLoader {
	return &WorldObjectLoader{
		archive: archive,
	}
}

func (l *WorldObjectLoader) Load() ([]*model.WorldObject, error) {
	// extract the files containing object data
	dataFile, err := l.archive.File("loc.dat")
	if err != nil {
		return nil, err
	}

	idxFile, err := l.archive.File("loc.idx")
	if err != nil {
		return nil, err
	}

	dataReader := asset.NewDataReader(dataFile)
	idxReader := asset.NewDataReader(idxFile)

	// read the number of objects in the file
	numObjects, err := idxReader.Uint16()
	if err != nil {
		return nil, err
	}

	objects := make([]*model.WorldObject, numObjects)

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

func (l *WorldObjectLoader) readObject(id int, r *asset.DataReader) (*model.WorldObject, error) {
	object := &model.WorldObject{
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
		case opObjectEndDefinition:
			// finished reading object definition
			hasMoreAttributes = false

		case opObjectModelIDsTypes:
			// read number of model
			numModels, err := r.Byte()
			if err != nil {
				return nil, err
			}

			_, err = r.Seek(int64(numModels*3), io.SeekCurrent)
			if err != nil {
				return nil, err
			}

		case opObjectName:
			// read the object's name
			name, err := r.String()
			if err != nil {
				return nil, err
			}

			object.Name = name

		case opObjectDescription:
			// read the object's description
			desc, err := r.String()
			if err != nil {
				return nil, err
			}

			object.Description = desc

		case opObjectModelIDs:
			// read number of model ids
			numModelIDs, err := r.Byte()
			if err != nil {
				return nil, err
			}

			_, err = r.Seek(int64(numModelIDs)*2, io.SeekCurrent)
			if err != nil {
				return nil, err
			}

		case opObjectSizeX:
			// read the object's size along the x-axis
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.Size.X = int(b)

		case opObjectSizeY:
			// read the object's size along the y-axis
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.Size.Y = int(b)

		case opObjectNotSolidFlag:
			// flag the object as solid
			object.Solid = false

		case opObjectNotWalkableFlag:
			// flag that the object cannot be walked on
			object.Walkable = false

		case opObjectActions:
			// read flag indicating if object has actions
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.HasActions = b == 0x01

		case opObjectTerrainFlag:
			object.AdjustToTerrain = true

		case opObjectShadingFlag:
			object.DelayedShading = true

		case opObjectWallFlag:
			object.Wall = true

		case opObjectAnimationID:
			// skip 2 bytes containing animation id
			_, err := r.Uint16()
			if err != nil {
				return nil, err
			}

		case opObjectOffsetAmpFlag:
			// read flag indicating if object has its offset amplified
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.OffsetAmplified = b == 0x01

		case opObjectAmbientFlag:
			// skip byte indicating if object has ambience
			_, err := r.Byte()
			if err != nil {
				return nil, err
			}

		case opObjectDiffuseFlag:
			// skip byte indicating if object has diffuse properties
			_, err := r.Byte()
			if err != nil {
				return nil, err
			}

		case opObjectColorCount:
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

		case opObjectIcon:
			// skip 2 bytes containing object icon
			_, err := r.Uint16()
			if err != nil {
				return nil, err
			}

		case opObjectRotatedFlag:
			object.Rotated = true

		case opObjectShadowFlag:
			object.Shadowless = true

		case opObjectScaleX:
			// read object's scale along x-axis
			scaleX, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Scale.X = int(scaleX)

		case opObjectScaleY:
			// read object's scale along y-axis
			scaleY, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Scale.Y = int(scaleY)

		case opObjectScaleZ:
			// read object's scale along z-axis
			scaleZ, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Scale.Z = int(scaleZ)

		case opObjectMapScene:
			// read object's map scene id
			mapSceneID, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.MapSceneID = int(mapSceneID)

		case opObjectFace:
			// read object's face id
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.FaceID = int(b)

		case opObjectTranslateX:
			// read object's translation along x-axis
			translateX, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Translation.X = int(translateX)

		case opObjectTranslateY:
			// read object's translation along y-axis
			translateY, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Translation.Y = int(translateY)

		case opObjectTranslateZ:
			// read object's translation along z-axis
			translateZ, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			object.Translation.Z = int(translateZ)

		case opObjectUnknown1:
			// unknown attribute
			break

		case opObjectUnwalkableFlag:
			object.UnwalkableSolid = true

		case opObjectStaticFlag:
			// read flag indicating if object is static
			b, err := r.Byte()
			if err != nil {
				return nil, err
			}

			object.Static = b == 0x01

		case opObjectEndAttributeList:
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
			if op >= opObjectActionListStart && op <= opObjectActionListEnd {
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
					idx := int(op - opObjectActionListStart)
					object.Actions[idx] = action
				}
			}
		}
	}

	return object, nil
}
